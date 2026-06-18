resource "noderush_volume" "data" {
  name        = "app-data"
  region_code = "fra"
  size_gb     = 50 # grow-only; increasing this resizes in place
}
