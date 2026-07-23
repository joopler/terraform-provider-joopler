// terraform-provider-joopler manages the scaffolding of a Joopler compliance
// program as code: control and policy ownership, vendors/subprocessors, the
// audit target, and connectors. It deliberately does NOT sign off on policies or
// attest evidence - those are human acts, kept out of the provider by design.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/joopler/terraform-provider-joopler/internal/provider"
)

// version is set at build time by goreleaser.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/joopler/joopler",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
