package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/EMSL-MSC/terraform-provider-ovirt/ovirt"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: ovirt.Provider,
	})
}
