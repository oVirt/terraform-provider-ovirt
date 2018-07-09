// Copyright (C) 2017 Battelle Memorial Institute
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

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
