// Copyright (C) 2017 Battelle Memorial Institute
// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"os"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4"
)

// Provider returns oVirt provider configuration
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVIRT_USERNAME", os.Getenv("OVIRT_USERNAME")),
				Description: "Login username",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVIRT_PASSWORD", os.Getenv("OVIRT_PASSWORD")),
				Description: "Login password",
				Sensitive:   true,
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVIRT_URL", os.Getenv("OVIRT_URL")),
				Description: "Ovirt server url",
			},
			"headers": {
				Type:        schema.TypeMap,
				Optional:    true,
				Default:     map[string]string{},
				Description: "Additional headers to be added to each API call",
			},
		},
		ConfigureFunc: ConfigureProvider,
		ResourcesMap: map[string]*schema.Resource{
			"ovirt_vm":              resourceOvirtVM(),
			"ovirt_disk":            resourceOvirtDisk(),
			"ovirt_disk_attachment": resourceOvirtDiskAttachment(),
			"ovirt_datacenter":      resourceOvirtDataCenter(),
			"ovirt_network":         resourceOvirtNetwork(),
			"ovirt_vnic":            resourceOvirtVnic(),
			"ovirt_vnic_profile":    resourceOvirtVnicProfile(),
			"ovirt_storage_domain":  resourceOvirtStorageDomain(),
			"ovirt_user":            resourceOvirtUser(),
			"ovirt_cluster":         resourceOvirtCluster(),
			"ovirt_mac_pool":        resourceOvirtMacPool(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"ovirt_disks":          dataSourceOvirtDisks(),
			"ovirt_datacenters":    dataSourceOvirtDataCenters(),
			"ovirt_networks":       dataSourceOvirtNetworks(),
			"ovirt_clusters":       dataSourceOvirtClusters(),
			"ovirt_storagedomains": dataSourceOvirtStorageDomains(),
			"ovirt_vnic_profiles":  dataSourceOvirtVNicProfiles(),
			"ovirt_authzs":         dataSourceOvirtAuthzs(),
			"ovirt_users":          dataSourceOvirtUsers(),
			"ovirt_mac_pools":      dataSourceOvirtMacPools(),
		},
	}
}

func castHeaders(h map[string]interface{}) map[string]string {
	headers := map[string]string{}

	for hk, hv := range h {
		headers[hk] = hv.(string)
	}

	return headers
}

// ConfigureProvider initializes the API connection object by config items
func ConfigureProvider(d *schema.ResourceData) (interface{}, error) {
	return ovirtsdk4.NewConnectionBuilder().
		URL(d.Get("url").(string)).
		Username(d.Get("username").(string)).
		Password(d.Get("password").(string)).
		Insecure(true).
		Headers(castHeaders(d.Get("headers").(map[string]interface{}))).
		Build()
}
