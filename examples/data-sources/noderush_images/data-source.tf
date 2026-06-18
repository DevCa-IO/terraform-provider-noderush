data "noderush_images" "all" {}

output "ubuntu_image_id" {
  value = one([for i in data.noderush_images.all.images : i.id if i.os == "ubuntu" && i.active])
}
