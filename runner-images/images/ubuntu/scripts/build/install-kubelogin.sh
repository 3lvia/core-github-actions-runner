#!/bin/bash -e
################################################################################
##  File:  install-kubelogin.sh
##  Desc:  Install kubelogin
################################################################################

# Don't install kubectl, install to /dev/null
az aks install-cli --install-location=/dev/null
