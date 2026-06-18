---
page_title: "noderush_ssh_key Resource - noderush"
subcategory: ""
description: |-
  A NodeRush SSH key, injected into Linux nodes at deploy time.
---

# noderush_ssh_key (Resource)

A NodeRush SSH key, injected into Linux nodes at deploy time. SSH keys are
immutable: changing the name or public key forces a new key.

## Example Usage

```terraform
resource "noderush_ssh_key" "deploy" {
  name       = "ci-deploy"
  public_key = file("~/.ssh/id_ed25519.pub")
}
```

## Schema

### Required

- `name` (String) Display name. Changing it forces a new key.
- `public_key` (String) The OpenSSH public key. Changing it forces a new key.

### Read-Only

- `id` (String) Opaque SSH key id.
- `fingerprint` (String) SHA256 fingerprint computed by the API.

## Import

Import is supported using the following syntax:

```shell
terraform import noderush_ssh_key.deploy ssh_abc123
```
