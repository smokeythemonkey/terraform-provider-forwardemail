<a href="https://terraform.io">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset=".github/terraform_logo_dark.svg">
    <source media="(prefers-color-scheme: light)" srcset=".github/terraform_logo_light.svg">
    <img src=".github/terraform_logo_light.svg" alt="Terraform logo" title="Terraform" align="right" height="50">
  </picture>
</a>

# Terraform Forward Email Provider

The [Forward Email Provider](https://registry.terraform.io/providers/smokeythemonkey/forwardemail/latest) enables [Terraform](https://terraform.io) to manage [Forward Email](https://forwardemail.net/) resources.

This has been forked from [forwardemail/terraform-provider-forwardemail](https://github.com/forwardemail/terraform-provider-forwardemail) due to lack of maintenance or activity for basic features like import, deploying sieve scripts etc.

This fork has updated the provider with:

- migration to terraform plugin-framework
- bumped forwardemail-api-go to 1.4.0
- added import functionality for domain and alias resources

Roadmap:

- Align provider with forwardemail settings
- Add sieve script resource

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0 (protocol 6; OpenTofu works too)
- [Go](https://golang.org/doc/install) >= 1.25 (to build the provider plugin)
- [mise](https://mise.jdx.dev/) (build toolchain; installs Go, Terraform, goreleaser, and linters per `mise.toml`)

## Building the Provider

Clone repository:

```sh
git clone https://github.com/smokeythemonkey/terraform-provider-forwardemail.git
```

Enter the provider directory and build the provider:

```sh
cd terraform-provider-forwardemail
mise run build
```

To use a released provider in your Terraform environment, run [`terraform init`](https://www.terraform.io/docs/commands/init.html) and Terraform will automatically install the provider. To specify a particular provider version when installing released providers, see the [Terraform documentation on provider versioning](https://www.terraform.io/docs/configuration/providers.html#version-provider-versions).

To instead use a custom-built provider in your Terraform environment (e.g. the provider binary from the build instructions above), follow the instructions to [install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-plugins) After placing the custom-built provider into your plugins directory, run `terraform init` to initialize it.

## Releasing

Releases are built, signed, and published by [goreleaser](https://goreleaser.com/) through mise, so the exact same commands run locally and in CI:

- `mise run release-check` — validate the `.goreleaser.yaml` configuration
- `mise run release-dry` — build the release artifacts locally without publishing
- `mise run release` — build, sign, and publish a release for the current tag

## Examples

```terraform
terraform {
  required_version = ">= 1.0.0"

  required_providers {
    forwardemail = {
      source  = "smokeythemonkey/forwardemail"
      version = ">= 0.1"
    }
  }
}

provider "forwardemail" {
  api_key = "XXXXXXX"
}

resource "forwardemail_domain" "default" {
  name = "example_domain.com"
}
```

## Contributing

For bug reports & feature requests, please use the [issue tracker](https://github.com/smokeythemonkey/terraform-provider-forwardemail/issues).

PRs are welcome! We follow the typical "fork-and-pull" Git workflow.

1. **Fork** the repo on GitHub
2. **Clone** the project to your own machine
3. **Commit** changes to your own branch
4. **Push** your work back up to your fork
5. Submit a **Pull Request** so that we can review your changes

> [!TIP]
> Be sure to merge the latest changes from "upstream" before making a pull request!

### Many Thanks to Our Contributors

<a href="https://github.com/smokeythemonkey/terraform-provider-forwardemail/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=smokeythemonkey/terraform-provider-forwardemail&max=24" />
</a>
