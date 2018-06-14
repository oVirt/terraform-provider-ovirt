// Copyright (C) 2017 Battelle Memorial Institute
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4"
)

// Provider returns oVirt provider configuration
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Login username",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Login password",
				Sensitive:   true,
			},
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Ovirt server url",
			},
		},
		ConfigureFunc: ConfigureProvider,
		ResourcesMap: map[string]*schema.Resource{
			"ovirt_vm":              resourceVM(),
			"ovirt_disk":            resourceDisk(),
			"ovirt_disk_attachment": resourceDiskAttachment(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"ovirt_disk":        dataSourceDisk(),
			"ovirt_datacenters": dataSourceOvirtDataCenters(),
		},
	}
}

func ConfigureProvider(d *schema.ResourceData) (interface{}, error) {
	return ovirtsdk4.NewConnectionBuilder().
		URL(d.Get("url").(string)).
		Username(d.Get("username").(string)).
		Password(d.Get("password").(string)).
		Insecure(true).
		Build()
}
