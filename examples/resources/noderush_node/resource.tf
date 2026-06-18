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
