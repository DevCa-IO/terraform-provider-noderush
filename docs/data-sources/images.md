---
page_title: "noderush_images Data Source - noderush"
subcategory: ""
description: |-
  The OS images you can deploy a node from.
---

# noderush_images (Data Source)

The OS images you can deploy a `noderush_node` from.

## Example Usage

```terraform
data "noderush_images" "all" {}

output "ubuntu_image_id" {
  value = one([for i in data.noderush_images.all.images : i.id if i.os == "ubuntu" && i.active])
}
```

## Schema

### Read-Only

- `images` (Attributes List) (see [below for nested schema](#nestedatt--images))

<a id="nestedatt--images"></a>
### Nested Schema for `images`

Read-Only:

- `id` (String) Image id to pass to `noderush_node.image_id`.
- `os` (String) OS family, e.g. `ubuntu`, `windows`.
- `label` (String) Human-readable label.
- `is_windows` (Boolean) True for Windows images.
- `active` (Boolean) Whether the image can be deployed.
