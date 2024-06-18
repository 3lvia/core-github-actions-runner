#!/bin/bash

set -eou pipefail

main() {
    if [[ -z "${ARM_RESOURCE_GROUP-}" ]]; then
        echo 'ARM_RESOURCE_GROUP is not set.'
        exit 1
    fi

    if [[ -z "${IMAGE_NAME-}" ]]; then
        echo 'IMAGE_NAME is not set.'
        exit 1
    fi

    local runners
    runners=$(az vmss list \
                --query "[?tags.* | [0] == 'github-actions'].name" \
                -o tsv)

    if [[ -z "$runners" ]]; then
        echo 'No runners found.'
        exit 0
    fi

    printf "Found runners: \n%s\n\n" "$runners"

    for runner in $runners; do
        local image_id
        image_id=$(az image show \
                    -g "${ARM_RESOURCE_GROUP}" \
                    -n "${IMAGE_NAME}" \
                    --query id \
                    -o tsv)

        local old_image_id
        old_image_id=$(az vmss show \
                        -g "${ARM_RESOURCE_GROUP}" \
                        -n "$runner" \
                        --query virtualMachineProfile.storageProfile.imageReference.id \
                        -o tsv)

        echo "Updating image for runner '$runner' to image '$image_id'..."
        az vmss update \
          -g "${ARM_RESOURCE_GROUP}" \
          -n "$runner" \
          --output none \
          --set virtualMachineProfile.storageProfile.imageReference.id="$image_id"
        printf "Done.\n\n"
    done

    sleep 5

    echo "Deleting old image '$old_image_id'..."
    az image delete \
       --output none \
       --ids "$old_image_id"
    echo 'Done.'
}

main
