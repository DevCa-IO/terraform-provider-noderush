data "noderush_plans" "fra" {
  region_code = "fra"
}

output "cheapest_fra_plan" {
  value = sort([for p in data.noderush_plans.fra.plans : p.monthly_cents])[0]
}
