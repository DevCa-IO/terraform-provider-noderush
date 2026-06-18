package main

import (
	"context"
	"flag"
	"log"

	"github.com/DevCa-IO/terraform-provider-noderush/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Run "go generate ./..." to format the examples and regenerate the docs/
// folder from the provider schema (requires the `terraform` binary on PATH).
//go:generate terraform fmt -recursive ./examples/
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name noderush

var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		// Published as registry.terraform.io/DevCa-IO/noderush once released.
		Address: "registry.terraform.io/DevCa-IO/noderush",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
