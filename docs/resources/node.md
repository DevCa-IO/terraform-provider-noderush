---
page_title: "noderush_node Resource - noderush"
subcategory: ""
description: |-
  A NodeRush VPS node. Create blocks until the node finishes provisioning.
---

# noderush_node (Resource)

A NodeRush VPS node. `terraform apply` blocks until the node finishes
provisioning (status `ONLINE`), up to 15 minutes. Deploying charges your wallet
upfront (first hour for `HOURLY`, first month for `MONTHLY`), so the workspace
must have sufficient balance.

Every configuration attribute forces replacement — in-place resize/rescale is
not yet supported.

## Example Usage

```terraform
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

  billing_mode = "HOURLY"
  ssh_key_ids  = [noderush_ssh_key.deploy.id]
}

output "web_ipv4" {
  value = noderush_node.web.ipv4
}
```

## Schema

### Required

- `hostname` (String) Hostname (<= 63 chars, DNS-label safe). Forces replacement.
- `region_code` (String) Region to deploy in (e.g. `fra`, `iad`). Forces replacement.
- `image_id` (String) OS image id (see the `noderush_images` data source). Forces replacement.
- `cpu` (Number) vCPU cores. Forces replacement.
- `ram_gb` (Number) Memory in GB. Forces replacement.
- `disk_gb` (Number) Disk in GB. Forces replacement.

### Optional

- `billing_mode` (String) `HOURLY` (default) or `MONTHLY`. Charged upfront at deploy. Forces replacement.
- `sku_id` (String) Optional SKU/plan id to pin pricing (see the `noderush_plans` data source). Forces replacement.
- `cloud_init` (String) Optional cloud-init script run on first boot. Forces replacement.
- `ssh_key_ids` (List of String) SSH key ids to inject (Linux images). Forces replacement.

### Read-Only

- `id` (String) Opaque node id.
- `ipv4` (String) Allocated IPv4 address.
- `ipv6` (String) Allocated IPv6 address, if any.
- `status` (String) Lifecycle status (`ONLINE` once provisioned).

## Import

Import is supported using the following syntax:

```shell
terraform import noderush_node.web node_abc123
```
