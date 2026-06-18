---
page_title: "noderush_plans Data Source - noderush"
subcategory: ""
description: |-
  The compute plans (SKUs) available, optionally filtered to a region.
---

# noderush_plans (Data Source)

The compute plans (SKUs) available, optionally filtered to a region. Use a
plan's `id` for `noderush_node.sku_id` to pin pricing.

## Example Usage

```terraform
data "noderush_plans" "fra" {
  region_code = "fra"
}

output "cheapest_fra_plan" {
  value = sort([for p in data.noderush_plans.fra.plans : p.monthly_cents])[0]
}
```

## Schema

### Optional

- `region_code` (String) Filter to plans available in this region.

### Read-Only

- `plans` (Attributes List) (see [below for nested schema](#nestedatt--plans))

<a id="nestedatt--plans"></a>
### Nested Schema for `plans`

Read-Only:

- `id` (String) Plan/SKU id.
- `family` (String) Plan family, e.g. `STANDARD`.
- `label` (String) Human-readable label.
- `cpu` (Number) vCPU cores.
- `ram_gb` (Number) Memory in GB.
- `disk_gb` (Number) Disk in GB.
- `hourly_cents` (Number) Hourly price in cents.
- `monthly_cents` (Number) Monthly price in cents.
