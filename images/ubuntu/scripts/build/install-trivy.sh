#!/bin/bash -e
################################################################################
##  File:  install-trivy.sh
##  Desc:  Install Trivy
################################################################################

# Source the helpers for use with the script
source $HELPER_SCRIPTS/install.sh

download_url=$(resolve_github_release_asset_url "aquasecurity/trivy" "startswith(\"trivy_\") and endswith(\"_Linux-64bit.tar.gz\")" "latest")
archive_path=$(download_with_retry "$download_url")

# TODO: check signature

tar -xzf "$archive_path" -C /usr/local/bin/ trivy
