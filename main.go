package main

import (
	"context"
	"flag"
	"log"

	"github.com/jamalshahverdiev/terraform-provider-freeswitch/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var version = "0.1.0"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run with support for debuggers")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/local/freeswitch",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
