provider "noderush" {
  # api_url defaults to https://api.noderush.io
  # api_token falls back to the NODERUSH_API_TOKEN environment variable.
  api_token = var.noderush_token
}
