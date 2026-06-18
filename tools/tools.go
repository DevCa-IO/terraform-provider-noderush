//go:build tools

// Tracks the tfplugindocs build dependency used to regenerate the docs/ folder
// from the provider schema + examples. Run `go generate ./...` (needs the
// `terraform` binary on PATH).
package tools

import (
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
