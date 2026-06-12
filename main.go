package main

import (
	"context"
	"flag"
	"log"

	"github.com/jamalshahverdiev/terraform-provider-freeswitch/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version is set by goreleaser at release build time (ldflags).
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run with support for debuggers")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/jamalshahverdiev/freeswitch",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
