// Copyright (C) 2017 Battelle Memorial Institute
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"fmt"
	"strconv"

	"github.com/EMSL-MSC/ovirtapi"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceDisk() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDiskRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"format": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"storage_domain_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"shareable": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"sparse": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func dataSourceDiskRead(d *schema.ResourceData, meta interface{}) error {
	con := meta.(*ovirtapi.Connection)
	disks, err := con.GetAllDisks()
	if err != nil {
		d.SetId("")
		return err
	}
	for _, disk := range disks {
		if disk.Name == d.Get("name") {
			d.Set("size", disk.ProvisionedSize)
			d.Set("format", disk.Format)
			d.Set("storage_domain_id", disk.StorageDomains.StorageDomain[0].ID)
			shareable, _ := strconv.ParseBool(disk.Shareable)
			d.Set("shareable", shareable)
			sparse, _ := strconv.ParseBool(disk.Sparse)
			d.Set("sparse", sparse)
			return nil
		}
	}

	return fmt.Errorf("Disk %s not found", d.Get("name"))
}
