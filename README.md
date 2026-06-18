# Terraform Provider for NodeRush

Manage NodeRush resources (block volumes, SSH keys) declaratively. The provider
talks to the NodeRush API gateway using a personal access token (PAT).

> Resources: `noderush_node`, `noderush_ssh_key`, `noderush_volume`.
> Data sources: `noderush_regions`, `noderush_images`, `noderush_plans`.
> `noderush_node` apply blocks until the node is provisioned (ONLINE) and bills
> the wallet upfront.

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
data "noderush_images" "all" {}

resource "noderush_ssh_key" "deploy" {
  name       = "ci-deploy"
  public_key = file("~/.ssh/id_ed25519.pub")
}

resource "noderush_node" "web" {
  hostname    = "web-1"
  region_code = "fra"
  image_id    = one([for i in data.noderush_images.all.images : i.id if i.os == "ubuntu" && i.active])
  cpu         = 2
  ram_gb      = 4
  disk_gb     = 80
  ssh_key_ids = [noderush_ssh_key.deploy.id]
}

resource "noderush_volume" "data" {
  name        = "app-data"
  region_code = "fra"
  size_gb     = 50 # grow-only; increasing this resizes in place
}

output "web_ipv4" {
  value = noderush_node.web.ipv4
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
