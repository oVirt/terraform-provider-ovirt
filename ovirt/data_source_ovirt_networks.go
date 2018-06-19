// Copyright (C) 2017 Battelle Memorial Institute
// Copyright (C) 2018 Boris Manojlovic
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4"
)

func dataSourceOvirtNetworks() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOvirtNetworksRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"networks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"datacenter_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vlan_id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"mtu": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}
}
func dataSourceOvirtNetworksRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)
	listResp, err := conn.SystemService().NetworksService().
		List().
		Search(fmt.Sprintf("name=%s", d.Get("name"))).
		Send()
	if err != nil {
		return err
	}

	networks, ok := listResp.Networks()
	if !ok && len(networks.Slice()) == 0 {
		d.SetId("")
		return fmt.Errorf("Network '%s' not found", d.Get("name"))
	}
	return networksDecriptionAttributes(d, networks.Slice(), meta)
}

func networksDecriptionAttributes(d *schema.ResourceData, network []*ovirtsdk4.Network, meta interface{}) error {
	var s []map[string]interface{}
	for _, v := range network {
		// description is not mandatory and if using MustDescription will fail with nil value
		desc, ok := v.Description()
		if !ok {
			desc = ""
		}
		mapping := map[string]interface{}{
			"id":            v.MustId(),
			"datacenter_id": v.MustDataCenter().MustId(),
			"description":   desc,
			"vlan_id":       v.MustVlan().MustId(),
			"mtu":           v.MustMtu(),
		}
		s = append(s, mapping)
	}
	d.SetId(resource.UniqueId())
	if err := d.Set("networks", s); err != nil {
		return err
	}

	return nil
}
