data "noderush_regions" "all" {}

output "fra_status" {
  value = one([for r in data.noderush_regions.all.regions : r.status if r.code == "fra"])
}
