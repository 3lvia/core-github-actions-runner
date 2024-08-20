#!/bin/bash -e
################################################################################
##  File:  install-playwright-dependencies.sh
##  Desc:  Install Playwright dependencies
################################################################################

echo 'Installing Playwright dependencies...'

apt install -y --no-install-recommends \
    libwoff1 \
    libvpx7 \
    libevent-2.1-7 \
    libopus0 \
    libgstreamer-plugins-base1.0-0 \
    libenchant-2-2 \
    libhyphen0 \
    libmanette-0.2-0

printf "Done.\n\n"
