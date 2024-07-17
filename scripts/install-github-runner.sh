#!/bin/bash -e
################################################################################
##  File:  install-github-runner.sh
##  Desc:  Install GitHub Runner
################################################################################

# Source the helpers for use with the script
# source $HELPER_SCRIPTS/install.sh
#
# TODO: use the helper functions from install.sh to download and install the GitHub Runner

gh_runner_version='2.317.0' # NOTE: to upgrade, also update the checksum below. See https://github.com/actions/runner/releases.
gh_runner_checksum='9e883d210df8c6028aff475475a457d380353f9d01877d51cc01a17b2a91161d'
gh_runner_tarball="actions-runner-linux-x64-$gh_runner_version.tar.gz"

echo 'Creating directory and downloading GitHub Runner...'
actions_dir='/opt/actions-runner'
mkdir -p "$actions_dir" && cd "$actions_dir"

curl -o "$gh_runner_tarball" \
    -L "https://github.com/actions/runner/releases/download/v$gh_runner_version/$gh_runner_tarball"
printf "Done.\n\n"

echo 'Verify GitHub Runner checksum...'
echo "$gh_runner_checksum  $gh_runner_tarball" | shasum -a 256 -c
printf "Done.\n\n"

echo 'Extracting GitHub Runner...'
tar xzf "./$gh_runner_tarball"
printf "Done.\n\n"
