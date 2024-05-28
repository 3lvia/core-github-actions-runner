# core-github-actions-runner

Configuration for generating the OS images used by Elvias GitHub Actions runners.

The image is generated in the GitHub workflow [generate-image.yml](.github/workflows/generate-image.yml) using Packer.
Packer will also push the image to Azure.

The Terraform code for the VM scale sets running the image is in the [core-terraform](https://github.com/3lvia/core-terraform) repository.
The credentials for authenticating to Azure are stored in the GitHub repository variables/secrets and are also configured in the core-terraform repository.

## Devlopment

We use two branches `master` and `develop` and their corresponding environments `prod` and `dev`.
This mirrors the setup in the core-terraform repository. This is to simplify the setup of credentials for GitHub Actions in core-terraform.
A consequence of this is also that we can test building images in the `develop` branch before merging to `master`.

The image produced by the `develop` branch is used by a single VMSS in the `dev` environment of core-terraform.
Use this for testing. SSH access to the VMSS (for dev only!) can also be enabled in core-terraform.
You need to supply your own SSH public key (RSA) in the module declaration.
