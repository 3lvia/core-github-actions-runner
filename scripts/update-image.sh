#!/bin/bash

set -eou pipefail

main() {
    local tmp_dir
    tmp_dir="$(mktemp -d)"

    local working_dir
    working_dir="$(pwd)"

    git clone git@github.com:actions/runner-images.git "$tmp_dir"

    mkdir -p images
    mkdir -p helpers

    cd "$tmp_dir"
    local files_images_ubuntu
    files_images_ubuntu=$(git ls-files 'images/ubuntu')

    local files_helpers
    files_helpers=$(git ls-files 'helpers')
    cd "$working_dir"

    local files_images_ubuntu_diff=()
    for file in $files_images_ubuntu; do
        local file_diff
        file_diff=$(diff --color "$tmp_dir/$file" "$working_dir/$file")
        printf "%s\n" "$file_diff"
        files_images_ubuntu_diff+=("$file_diff")
    done

    if [[ ${#files_images_ubuntu_diff[@]} -gt 0 ]]; then
        read -p "Do you want to update the images/ubuntu files? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            for file in $files_images_ubuntu; do
                printf "Copying %s\n" "$file"
                cp "$tmp_dir/$file" "$working_dir/$file"
            done
        fi
    else
        echo "No changes in images/ubuntu files"
    fi

    local files_helpers_diff=()
    for file in $files_helpers; do
        local file_diff
        file_diff=$(diff --color "$tmp_dir/$file" "$working_dir/$file")
        printf "%s\n" "$file_diff"
        files_helpers_diff+=("$file_diff")
    done

    if [[ ${#files_helpers_diff[@]} -gt 0 ]]; then
        read -p "Do you want to update the helpers files? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            for file in $files_helpers; do
                printf "Copying %s\n" "$file"
                cp "$tmp_dir/$file" "$working_dir/$file"
            done
        fi
    else
        echo "No changes in helpers files"
    fi

    rm -rf "$tmp_dir"
}

main
