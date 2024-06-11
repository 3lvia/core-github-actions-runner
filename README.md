# core-github-actions-runner

Configuration for generating the OS images used by Elvias GitHub Actions runners, called 'elvia-runner'.

The image is generated in the GitHub workflow [generate-image.yml](.github/workflows/generate-image.yml) using Packer.
Packer will also publish the image to Azure.

The Terraform code for the VM scale sets running the image is in the [core-terraform](https://github.com/3lvia/core-terraform) repository.
The credentials for authenticating to Azure are stored in GitHub environment variables/secrets and are also configured in [core-terraform](https://github.com/3lvia/core-terraform).

## Updating the image

### Syncing with upstream

The configuration for the image is based on the GitHub [runner-images](https://github.com/actions/runner-images) repository.
We have copied the configuration for the Ubuntu 22.04 image, and made some modifications.

Syncing should be handled automatically by the GitHub workflow [sync-upstream.yml](.github/workflows/sync-upstream.yml).
This workflow will run on a schedule and check for changes in the upstream repository, and create a pull request if necessary.
The pull request should include a short guide on how to proceed with merging the changes.

If you want to manually sync with the upstream repository, you can run the following command:

```bash
go run main.go
```

Packer and git must be installed on your machine for this to work.

### Remove software

We remove some software from the image to reduce build times.
To remove software from the image, edit the `remove_software_list` variable in [main.go](main.go).
After editing the file, run the following command:

```bash
go run main.go --apply
```

This should remove the required configuration from Packer and also remove the installation script.
Packer and git must be installed on your machine for this to work.
You can double check by checking the git diff.

### Add software

To add software to the image, edit the `add_software_list` variable in [main.go](main.go).

You will also need to supply an installation script in the [scripts](scripts) directory.
See [scripts/install-trivy.sh](scripts/install-trivy.sh) for an example.
Your script **MUST** follow the same naming convention, i.e.: `install-<software>.sh`.

As with removing software, run the following command:

```bash
go run main.go --apply
```

This should add the required configuration to Packer and copy the installation script to the correct location.
Packer and git must be installed on your machine for this to work.
You can double check by checking the git diff.

## Deleting old runners

When scaling the amount of VM instances, the VM's that get terminated will not be deregistered as GitHub runners.
This means that they will still show up in the list of runners in the organizations settings, but will be permanently 'offline'.
When registering new runners, they will usually replace the old ones since the hostname is the same.
However, this is not always the case. Therefore, we have a workflow that will delete all runners that are offline.
This is done in the [delete-runners.yml](.github/workflows/delete-runners.yml) workflow.

## Development

We use trunk-based development, and two environments `prod` and `dev`.
Any pull request to the trunk branch `trunk` will generate and push an image to the `dev` environment.
After merging to the `trunk` branch, the image will be pushed to the `prod` environment.
In both cases, these images will be deployed to either the VMSS '**elvia-runner-dev**' or the VMSS '**elvia-runner-prod**', respectively.

When testing, open a pull request to trunk and generate your image, which will be deployed to the `dev` environment.
SSH access is enabled for the dev environment, so you can connect to the VM and test your changes.
SSH access is configured in [core-terraform](https://github.com/3lvia/core-terraform), and you must provide your own public key.
