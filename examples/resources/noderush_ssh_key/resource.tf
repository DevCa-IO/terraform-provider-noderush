resource "noderush_ssh_key" "deploy" {
  name       = "ci-deploy"
  public_key = file("~/.ssh/id_ed25519.pub")
}
