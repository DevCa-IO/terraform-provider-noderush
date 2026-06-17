# Terraform Provider for NodeRush

Manage NodeRush resources (block volumes, SSH keys) declaratively. The provider
talks to the NodeRush API gateway using a personal access token (PAT).

> Status: initial release. Resources: `noderush_ssh_key`, `noderush_volume`.
> Data source: `noderush_regions`. The `noderush_node` resource is next (node
> deploys are asynchronous and bill upfront, so they need extra plan handling).

## Build

```sh
cd terraform-provider-noderush
go mod tidy
go build .
```

To use the locally built binary, add a dev override to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "DevCa-IO/noderush" = "/absolute/path/to/terraform-provider-noderush"
  }
  direct {}
}
```

## Configure

```hcl
terraform {
  required_providers {
    noderush = {
      source = "DevCa-IO/noderush"
    }
  }
}

provider "noderush" {
  # api_url defaults to https://api.noderush.io
  # both fall back to env vars NODERUSH_API_URL / NODERUSH_API_TOKEN
  api_token = var.noderush_token
}
```

Create the token in the dashboard under Settings → API keys.

## Example

```hcl
data "noderush_regions" "all" {}

resource "noderush_ssh_key" "deploy" {
  name       = "ci-deploy"
  public_key = file("~/.ssh/id_ed25519.pub")
}

resource "noderush_volume" "data" {
  name        = "app-data"
  region_code = "fra"
  size_gb     = 50 # grow-only; increasing this resizes in place
}

output "fra_status" {
  value = one([for r in data.noderush_regions.all.regions : r.status if r.code == "fra"])
}
```

## Publishing

### Public Terraform Registry

1. Move this directory into its own public GitHub repo named exactly
   `terraform-provider-noderush` (the registry requires that name at repo root).
2. Generate a GPG key; add the **public** key to your Terraform Registry account
   (Settings → Signing Keys), and add `GPG_PRIVATE_KEY` (ASCII-armored) +
   `PASSPHRASE` as repo Actions secrets.
3. Connect the repo on registry.terraform.io and click Publish.
4. Push a semver tag: `git tag v0.1.0 && git push origin v0.1.0`. The included
   `.github/workflows/release.yml` runs GoReleaser, which builds the
   multi-platform zips, `_SHA256SUMS`, and the GPG signature; the registry
   ingests the GitHub release automatically. Consumers then use
   `source = "DevCa-IO/noderush"`.

### Private use (no public listing)

- `dev_overrides` (above) for local development.
- A **filesystem mirror**: `goreleaser release --snapshot --clean`, then drop the
  zips under `~/.terraform.d/plugins/registry.terraform.io/DevCa-IO/noderush/<version>/<os>_<arch>/`
  or point Terraform CLI at a mirror dir with `provider_installation { filesystem_mirror { path = "..." } }`.
- A Terraform Cloud/Enterprise **private registry**.

## Notes

- `noderush_volume` resize is grow-only; the API rejects a smaller `size_gb`,
  so the provider only calls resize when the size increases. Changing `name` or
  `region_code` forces replacement.
- `noderush_ssh_key` is immutable; any change forces a new key.
- Both resources support `terraform import <addr> <id>`.
