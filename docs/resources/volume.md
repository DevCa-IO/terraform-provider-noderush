---
page_title: "noderush_volume Resource - noderush"
subcategory: ""
description: |-
  A NodeRush block storage volume. Billed per GB-month. Resize is grow-only.
---

# noderush_volume (Resource)

A NodeRush block storage volume. Billed per GB-month. Resize is grow-only: you
can increase `size_gb` in place, but decreasing it forces a new volume.
Changing `name` or `region_code` also forces replacement.

## Example Usage

```terraform
resource "noderush_volume" "data" {
  name        = "app-data"
  region_code = "fra"
  size_gb     = 50 # grow-only; increasing this resizes in place
}
```

## Schema

### Required

- `name` (String) Display name (changing it forces a new volume).
- `region_code` (String) Region the volume lives in (e.g. `fra`, `iad`). Immutable.
- `size_gb` (Number) Size in GB. Can be increased in place; decreasing forces a new volume.

### Read-Only

- `id` (String) Opaque volume id.
- `status` (String) Lifecycle status.

## Import

Import is supported using the following syntax:

```shell
terraform import noderush_volume.data vol_abc123
```
