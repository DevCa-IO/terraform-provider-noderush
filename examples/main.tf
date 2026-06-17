terraform {
  required_providers {
    noderush = {
      source = "DevCa-IO/noderush"
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

data "noderush_regions" "all" {}

resource "noderush_ssh_key" "deploy" {
  name       = "ci-deploy"
  public_key = file("~/.ssh/id_ed25519.pub")
}

resource "noderush_volume" "data" {
  name        = "app-data"
  region_code = "fra"
  size_gb     = 50
}
