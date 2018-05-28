// Copyright (C) 2017 Battelle Memorial Institute
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4"
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
	conn := meta.(*ovirtsdk4.Connection)

	listResp, err := conn.SystemService().DisksService().
		List().Search(fmt.Sprintf("name=%s", d.Get("name"))).Send()
	if err != nil {
		return err
	}

	disks, ok := listResp.Disks()
	if !ok && len(disks.Slice()) > 0 {
		d.SetId("")
		return nil
	}

	disk := disks.Slice()[0]
	d.Set("size", disk.MustProvisionedSize())
	d.Set("format", disk.MustFormat())
	d.Set("storage_domain_id", disk.MustStorageDomains().Slice()[0].MustId())
	d.Set("shareable", disk.MustShareable())
	d.Set("sparse", disk.MustSparse())
	return nil
}
