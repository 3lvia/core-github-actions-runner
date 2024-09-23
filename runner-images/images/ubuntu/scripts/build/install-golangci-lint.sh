#!/bin/bash -e
################################################################################
##  File:  install-golangci-lint
##  Desc:  Install golangci-lint
################################################################################

VERSION='1.60.3'

echo 'Installing golangci-lint...'

# binary will be $(go env GOPATH)/bin/golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v${VERSION}

printf "Done.\n\n"
