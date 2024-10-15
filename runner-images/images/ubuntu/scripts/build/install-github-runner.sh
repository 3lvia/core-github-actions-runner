#!/bin/bash -e
################################################################################
##  File:  install-github-runner.sh
##  Desc:  Install GitHub Runner
################################################################################

echo 'Creating directory and downloading latest GitHub Runner...'
gh_actions_dir="/opt/actions-runner"
mkdir -p "$gh_actions_dir" && cd "$gh_actions_dir"

gh_runner_latest_version=$(curl -sL https://api.github.com/repos/actions/runner/releases/latest | jq -r .tag_name)
gh_runner_download_url=$(curl -sL "https://api.github.com/repos/actions/runner/releases/tags/$gh_runner_latest_version" | jq -r '.assets[] | select(.name | contains("linux-x64")) | .browser_download_url')
gh_runner_tarball=$(basename "$gh_runner_download_url")
gh_runner_checksum=$(curl -sL "https://api.github.com/repos/actions/runner/releases/tags/$gh_runner_latest_version" | jq -r .body | grep -oP '(?<=<!-- BEGIN SHA linux-x64 -->).*?(?=<!-- END SHA linux-x64 -->)')

curl -o "$gh_runner_tarball" -L "$gh_runner_download_url"
printf "Done.\n\n"

echo 'Verify GitHub Runner checksum...'
echo "$gh_runner_checksum  $gh_runner_tarball" | shasum -a 256 -c
printf "Done.\n\n"

echo 'Extracting GitHub Runner...'
tar xzf "./$gh_runner_tarball"
printf "Done.\n\n"
