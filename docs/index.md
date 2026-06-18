---
page_title: "noderush Provider"
description: |-
  Manage NodeRush resources (VPS nodes, block volumes, SSH keys) through the NodeRush API.
---

# noderush Provider

Manage NodeRush resources declaratively. The provider talks to the NodeRush API
gateway using a personal access token (PAT). Create a token in the dashboard
under Settings -> API keys.

## Example Usage

```terraform
provider "noderush" {
  # api_url defaults to https://api.noderush.io
  # api_token falls back to the NODERUSH_API_TOKEN environment variable.
  api_token = var.noderush_token
}
```

## Schema

### Optional

- `api_url` (String) Base URL of the NodeRush API. Defaults to the `NODERUSH_API_URL` env var, or `https://api.noderush.io`.
- `api_token` (String, Sensitive) A NodeRush personal access token. Defaults to the `NODERUSH_API_TOKEN` env var. Required (provide here or via the env var).
