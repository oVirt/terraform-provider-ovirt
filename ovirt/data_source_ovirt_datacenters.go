// Copyright (C) 2017 Battelle Memorial Institute
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

func dataSourceOvirtDataCenters() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOvirtDataCentersRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// Computed
			"datacenters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"local": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"quota_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceOvirtDataCentersRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	dcsResp, err := conn.SystemService().DataCentersService().
		List().
		Search(fmt.Sprintf("name=%s", d.Get("name").(string))).
		Send()
	if err != nil {
		return err
	}
	dcs, ok := dcsResp.DataCenters()
	if !ok || len(dcs.Slice()) == 0 {
		return fmt.Errorf("your query datacenter returned no results, please change your search criteria and try again")
	}

	return dataCentersDecriptionAttributes(d, dcs.Slice(), meta)
}

func dataCentersDecriptionAttributes(d *schema.ResourceData, dcs []*ovirtsdk4.DataCenter, meta interface{}) error {
	var s []map[string]interface{}
	for _, v := range dcs {
		mapping := map[string]interface{}{
			"id":         v.MustId(),
			"status":     v.MustStatus(),
			"local":      v.MustLocal(),
			"quota_mode": v.MustQuotaMode(),
		}
		s = append(s, mapping)
	}
	d.SetId(resource.UniqueId())
	if err := d.Set("datacenters", s); err != nil {
		return err
	}

	return nil
}
