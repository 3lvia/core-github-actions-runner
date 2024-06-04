# core-github-actions-runner

Configuration for generating the OS images used by Elvias GitHub Actions runners.

The image is generated in the GitHub workflow [generate-image.yml](.github/workflows/generate-image.yml) using Packer.
Packer will also publish the image to Azure.

The Terraform code for the VM scale sets running the image is in the [core-terraform](https://github.com/3lvia/core-terraform) repository.
The credentials for authenticating to Azure are stored in GitHub environment variables/secrets and are also configured in the core-terraform repository.

## Updating the image

### Syncing with upstream

The configuration for the image is based on the GitHub [runner-images](https://github.com/actions/runner-images) repository.
We have copied the configuration for the Ubuntu 22.04 image, and made some modifications.

Syncing should be handled automatically by the GitHub workflow [sync-upstream.yml](.github/workflows/sync-upstream.yml).
This workflow will run on a schedule and check for changes in the upstream repository, and create a pull request if necessary.

If you want to manually sync with the upstream repository, you can run the following command:

```bash
go run main.go
```

Git and Packer must be installed on your machine for this to work.

### Remove software

We remove some software from the image to reduce build times.
To remove software from the image, edit the `remove_software_list` variable in [main.go](main.go).
After editing the file, run the following command:

```bash
go run main.go --apply
```

This should remove the required configuration from Packer and also remove the installation script.
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
You can double check by checking the git diff.

## Development

We use two branches `master` and `develop` and their corresponding environments `prod` and `dev`.
This mirrors the setup in the core-terraform repository. This is to simplify the setup of credentials for GitHub Actions in core-terraform.
A consequence of this is also that we can test building images in the `develop` branch before merging to `master`.

The image produced by the `develop` branch is used by a single VMSS in the `dev` environment of core-terraform.
Use this for testing. SSH access to the VMSS (for dev only!) can also be enabled in core-terraform.
You need to supply your own SSH public key (RSA) in the module declaration.
