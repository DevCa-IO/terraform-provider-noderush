---
page_title: "noderush_regions Data Source - noderush"
subcategory: ""
description: |-
  The set of NodeRush regions you can deploy into.
---

# noderush_regions (Data Source)

The set of NodeRush regions you can deploy into.

## Example Usage

```terraform
data "noderush_regions" "all" {}

output "fra_status" {
  value = one([for r in data.noderush_regions.all.regions : r.status if r.code == "fra"])
}
```

## Schema

### Read-Only

- `regions` (Attributes List) (see [below for nested schema](#nestedatt--regions))

<a id="nestedatt--regions"></a>
### Nested Schema for `regions`

Read-Only:

- `code` (String) Region code, e.g. `fra`.
- `label` (String) Human-readable label.
- `country_code` (String) ISO country code.
- `status` (String) Operational status.
