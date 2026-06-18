---
page_title: "noderush Provider"
description: |-
  Manage NodeRush VPS nodes, block storage volumes, and SSH keys with Terraform.
---

# NodeRush Provider

[NodeRush](https://noderush.io) is a VPS and Windows-RDP hosting platform with
regions across the EU (Frankfurt, Amsterdam, Chișinău, Paris) and the USA
(Ashburn). This provider lets you manage NodeRush resources declaratively
through the NodeRush API.

It manages:

- **`noderush_node`** — a VPS node. `apply` blocks until it is provisioned.
- **`noderush_ssh_key`** — SSH keys injected into Linux nodes at deploy time.
- **`noderush_volume`** — block storage volumes (per-GB-month, grow-only resize).

Data sources:

- **`noderush_regions`** — the regions you can deploy into.
- **`noderush_images`** — the OS images for a node.
- **`noderush_plans`** — the compute plans (SKUs), optionally per region.

## Authentication

The provider authenticates with a **personal access token (PAT)**. Create one in
the NodeRush dashboard under **Settings → API keys**, then provide it via the
`api_token` argument or the `NODERUSH_API_TOKEN` environment variable. The token
scopes to the workspace it was created in.

```terraform
provider "noderush" {
  api_token = var.noderush_token # or set NODERUSH_API_TOKEN
}
```

## Example Usage

```terraform
terraform {
  required_providers {
    noderush = {
      source  = "DevCa-IO/noderush"
      version = "~> 0.1"
    }
  }
}

variable "noderush_token" {
  type      = string
  sensitive = true
}

provider "noderush" {
  api_token = var.noderush_token
}

data "noderush_images" "all" {}

# An SSH key to inject into Linux nodes.
resource "noderush_ssh_key" "deploy" {
  name       = "ci-deploy"
  public_key = file("~/.ssh/id_ed25519.pub")
}

# A VPS node in Frankfurt. apply blocks until it is ONLINE.
resource "noderush_node" "web" {
  hostname    = "web-1"
  region_code = "fra"
  image_id    = one([for i in data.noderush_images.all.images : i.id if i.os == "ubuntu" && i.active])
  cpu         = 2
  ram_gb      = 4
  disk_gb     = 80
  ssh_key_ids = [noderush_ssh_key.deploy.id]
}

# A 50 GB block volume in Frankfurt (grow-only).
resource "noderush_volume" "data" {
  name        = "app-data"
  region_code = "fra"
  size_gb     = 50
}

output "web_ipv4" {
  value = noderush_node.web.ipv4
}
```

## Schema

### Optional

- `api_url` (String) Base URL of the NodeRush API. Defaults to the `NODERUSH_API_URL` env var, or `https://api.noderush.io`.
- `api_token` (String, Sensitive) A NodeRush personal access token. Defaults to the `NODERUSH_API_TOKEN` env var. Required (provide here or via the env var).
