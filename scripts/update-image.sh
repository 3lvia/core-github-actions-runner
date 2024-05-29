#!/bin/bash

set -eou pipefail

check_diff() {
    local files_path="$1"
    local local_git_dir="$2"
    local git_dir="$3"

    local files

    cd "$git_dir"
    files=$(git ls-files "$files_path")
    cd "$local_git_dir"

    for file in $files; do
        local file_diff
        file_diff=$(diff -u --color "$local_git_dir/$file" "$git_dir/$file" || true)
        if [[ "$file_diff" != "" ]]; then
            printf "\n\n\nChanges in '%s':\n\n\n" "$file"
            echo "$file_diff"
            printf "\n\n\n"
            read -p "Do you want to apply these changes to $file? [y/N] " -n 1 -r
            printf "\n\n\n"
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                cp "$git_dir/$file" "$local_git_dir/$file"
            fi
        else
            printf "No changes in '%s'.\n" "$file"
        fi
    done
}

remove_software() {
    local template_file_rel="$1"
    local git_dir="$2"
    local template_file="$git_dir/$template_file_rel"
    local toolset_file="$git_dir/images/ubuntu/toolsets/toolset-2204.json"
    local remove_software_list=(
        'apache'
        'aws-tools'
        'gfortran'
        'java-tools'
        'php'
        'postgresql'
        'pulumi'
        'bazel'
        'rust'
        'julia'
        'selenium'
        'vcpkg'
        'android-sdk'
        'leiningen'
        'kotlin'
        'sbt'
        'oc-cli'
        'aliyun-cli'
        'rlang'
        'heroku'
    )

    printf "Disable software report generation...\n\n"

    local software_gen_block='  provisioner "shell" {
    environment_vars = ["IMAGE_VERSION=${var.image_version}", "INSTALLER_SCRIPT_FOLDER=${var.installer_script_folder}"]
    inline           = ["pwsh -File ${var.image_folder}/SoftwareReport/Generate-SoftwareReport.ps1 -OutputDirectory ${var.image_folder}", "pwsh -File ${var.image_folder}/tests/RunAll-Tests.ps1 -OutputDirectory ${var.image_folder}"]
  }

  provisioner "file" {
    destination = "${path.root}/../Ubuntu2204-Readme.md"
    direction   = "download"
    source      = "${var.image_folder}/software-report.md"
  }

  provisioner "file" {
    destination = "${path.root}/../software-report.json"
    direction   = "download"
    source      = "${var.image_folder}/software-report.json"
  }'

    printf "Please delete this block of code from '%s' (if it exists):\n\n%s\n\n" "$template_file" "$software_gen_block"
    read -p "Press any key when done." -n 1 -r
    printf "\n\n\n"

    printf "Removing software...\n\n"

    for software in "${remove_software_list[@]}"; do
        printf "    Removing install script for '%s'...\n" "$software"
        rm -f "$git_dir/images/ubuntu/scripts/build/install-$software.sh"

        printf "    Removing line from Packer configuration for '%s'...\n\n" "$software"
        sed -i "/install-$software.sh/d" "$template_file"

        # Special case for 'android-sdk' since toolset file uses 'android' instead
        if [[ "$software" == "android-sdk" ]]; then
            software='android'
        fi

        if grep -q "$software" "$toolset_file"; then
            printf "    Removing configuration from '%s' for '%s'...\n\n" "$toolset_file" "$software"
            sed -i '/    "'"$software"'":/,/    },/d' "$toolset_file"
        fi
    done

    printf "Done.\n\n"

    validate_packer "$template_file"
}

add_software() {
    local template_file_rel="$1"
    local local_dir="$2"
    local git_dir="$3"
    local template_file="$git_dir/$template_file_rel"
    local add_software_list=(
        'trivy'
    )

    printf "Adding software...\n\n"

    for software in "${add_software_list[@]}"; do
        local install_script="$git_dir/images/ubuntu/scripts/build/install-$software.sh"
        if [[ -f "$install_script" ]]; then
            printf "    Install script for '%s' already exists.\n" "$software"
        else
            printf "    Adding install script for '%s'...\n" "$software"
            cp "$local_dir/scripts/install-$software.sh" "$install_script"
        fi

        if grep -q "install-$software.sh" "$template_file"; then
            printf "    Line for '%s' already exists in Packer configuration.\n\n" "$software"
        else
            printf "    Adding line to Packer configuration for '%s'...\n\n" "$software"
            sed -i \
                's/"${path.root}\/\.\.\/scripts\/build\/install-zstd\.sh",*/"${path.root}\/\.\.\/scripts\/build\/install-zstd\.sh",\n      "${path.root}\/\.\.\/scripts\/build\/install-'"$software"'\.sh",/' \
                "$template_file"
        fi
    done

    printf "Done.\n\n"

    validate_packer "$template_file"
}

validate_packer() {
    local template_file="$1"

    echo 'Validating Packer configuration...'
    packer validate \
        -var managed_image_resource_group_name='test' \
        -var location='westeurope' \
        "$template_file"
    echo
}

apply_customizations() {
    local template_file="$1"
    local local_dir="$2"
    local git_dir="$3"

    remove_software "$template_file" "$git_dir"
    add_software "$template_file" "$local_dir" "$git_dir"
}

update() {
    local template_file="$1"
    local local_dir="$2"

    local git_dir
    git_dir="$(mktemp -d)/runner-images"

    echo 'Cloning runner-images repository...'
    git clone git@github.com:actions/runner-images.git "$git_dir" -q
    cd "$git_dir"
    printf "Done.\n\n"

    local latest_tag
    latest_tag=$(git tag | grep 'ubuntu22' | tail -n1)
    echo "Checking out release $latest_tag."
    git checkout "$latest_tag" -q
    cd "$local_dir"

    apply_customizations "$template_file" "$local_dir" "$git_dir"

    local files_dirs=(
        'images/ubuntu'
        'helpers'
    )

    for file_dir in "${files_dirs[@]}"; do
        local local_git_dir="$local_dir/runner-images"
        mkdir -p "$local_git_dir/$file_dir"
        check_diff "$file_dir" "$local_git_dir" "$git_dir"
    done

    rm -rf "$git_dir"

    echo
    validate_packer "$template_file"
    echo 'Done.'
}

main() {
    local template_file='images/ubuntu/templates/ubuntu-22.04.pkr.hcl'

    local local_dir
    local_dir="$(pwd)"

    if [[ "${1-}" == "--apply" ]]; then
        local git_dir="$local_dir/runner-images"
        apply_customizations "$template_file" "$local_dir" "$git_dir"
    else
        update "$template_file" "$local_dir"
    fi
}

main "$@"
