#!/bin/bash

set -eou pipefail

check_diff() {
    local files_path="$1"
    local git_dir="$2"
    local working_dir="$3"

    local files

    cd "$git_dir"
    files=$(git ls-files "$files_path")
    cd "$working_dir"

    for file in $files; do
        local file_diff
        file_diff=$(diff -u --color "$working_dir/$file" "$git_dir/$file" || true)
        if [[ "$file_diff" != "" ]]; then
            printf "\n\n\nChanges in '%s':\n\n\n" "$file"
            echo "$file_diff"
            printf "\n\n\n"
            read -p "Do you want to apply these changes to $file? [y/N] " -n 1 -r
            printf "\n\n\n"
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                cp "$git_dir/$file" "$working_dir/$file"
            fi
        else
            printf "No changes in '%s'.\n" "$file"
        fi
    done
}

remove_software() {
    local template_file="$1"
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
    )

    printf "Removing software...\n\n"

    for software in "${remove_software_list[@]}"; do
        printf "    Removing install script for '%s'...\n" "$software"
        rm -f "images/ubuntu/scripts/build/install-$software.sh"

        printf "    Removing line from Packer configuration for '%s'...\n\n" "$software"
        sed -i "/install-$software.sh/d" "$template_file"
    done

    printf "Done.\n\n"

    validate_packer "$template_file"
}

# TODO: Implement this function
add_software() {
    local template_file="$1"

    echo "Adding software..."
    echo 'NOT IMPLEMENTED'
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

    remove_software "$template_file"
    add_software "$template_file"
}

update() {
    local git_dir
    git_dir="$(mktemp -d)"

    local working_dir
    working_dir="$(pwd)"

    local template_file="$1"

    echo 'Cloning runner-images repository...'
    git clone git@github.com:actions/runner-images.git "$git_dir" -q
    cd "$git_dir"
    printf "Done.\n\n"

    apply_customizations "$template_file"

    local dirs='images/ubuntu helpers'
    cd "$working_dir"

    for dir_ in $dirs; do
        mkdir -p "$dir_"
        check_diff "$dir_" "$git_dir" "$working_dir"
    done

    rm -rf "$git_dir"
}

main() {
    local template_file='images/ubuntu/templates/ubuntu-22.04.pkr.hcl'

    if [[ "${1-}" == "--apply" ]]; then
        apply_customizations "$template_file"
    else
        update "$template_file"
    fi
}

main "$@"
