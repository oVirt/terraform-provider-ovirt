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

func dataSourceOvirtDisks() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOvirtDisksRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			// Computed
			"disks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"alias": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"format": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"storage_domain_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"shareable": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"sparse": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceOvirtDisksRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	listResp, err := conn.SystemService().DisksService().
		List().Search(fmt.Sprintf("name=%s", d.Get("name"))).Send()
	if err != nil {
		return err
	}

	disks, ok := listResp.Disks()
	if !ok && len(disks.Slice()) == 0 {
		return fmt.Errorf("your query disk returned no results, please change your search criteria and try again")
	}

	return disksDescriptionAttributes(d, disks.Slice(), meta)

}

func disksDescriptionAttributes(d *schema.ResourceData, disks []*ovirtsdk4.Disk, meta interface{}) error {
	var s []map[string]interface{}

	for _, v := range disks {
		mapping := map[string]interface{}{
			"id":     v.MustId(),
			"format": v.MustFormat(),
			"size":   v.MustProvisionedSize(),
		}
		if sds, ok := v.StorageDomains(); ok {
			if len(sds.Slice()) > 0 {
				mapping["storage_domain_id"] = sds.Slice()[0].MustId()
			}
		}
		if alias, ok := v.Alias(); ok {
			mapping["alias"] = alias
		}
		if shareable, ok := v.Shareable(); ok {
			mapping["shareable"] = shareable
		}
		if sparse, ok := v.Sparse(); ok {
			mapping["sparse"] = sparse
		}

		s = append(s, mapping)
	}
	d.SetId(resource.UniqueId())
	if err := d.Set("disks", s); err != nil {
		return err
	}

	return nil
}
