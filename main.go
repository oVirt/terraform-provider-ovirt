package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/ovirt/terraform-provider-ovirt/ovirt"
)

func main() {
	var debugMode bool

	flag.BoolVar(
		&debugMode,
		"debug",
		false,
		"set to true to run the provider with support for debuggers like delve",
	)
	flag.Parse()

	opts := &plugin.ServeOpts{
		ProviderFunc: ovirt.New(),
	}

	if debugMode {
		err := plugin.Debug(
			context.Background(),
			"registry.terraform.io/ovirt/ovirt",
			opts,
		)
		if err != nil {
			log.Fatal(err.Error())
		}
		return
	}

	plugin.Serve(opts)
}
