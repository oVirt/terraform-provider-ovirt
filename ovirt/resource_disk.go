// Copyright (C) 2017 Battelle Memorial Institute
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"github.com/hashicorp/terraform/helper/schema"
	ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4"
)

func resourceDisk() *schema.Resource {
	return &schema.Resource{
		Create: resourceDiskCreate,
		Read:   resourceDiskRead,
		Delete: resourceDiskDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"format": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"storage_domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"shareable": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"sparse": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			// "qcow_version" is the only field supporting Disk-Update
		},
	}
}

func resourceDiskCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	diskBuilder := ovirtsdk4.NewDiskBuilder().
		Name(d.Get("name").(string)).
		Format(ovirtsdk4.DiskFormat(d.Get("format").(string))).
		ProvisionedSize(int64(d.Get("size").(int))).
		StorageDomainsOfAny(
			ovirtsdk4.NewStorageDomainBuilder().
				Id(d.Get("storage_domain_id").(string)).
				MustBuild())
	if shareable, ok := d.GetOkExists("shareable"); ok {
		diskBuilder.Shareable(shareable.(bool))
	}
	if sparse, ok := d.GetOkExists("sparse"); ok {
		diskBuilder.Sparse(sparse.(bool))
	}
	disk, err := diskBuilder.Build()
	if err != nil {
		return err
	}

	addResp, err := conn.SystemService().DisksService().Add().Disk(disk).Send()
	if err != nil {
		return err
	}

	d.SetId(addResp.MustDisk().MustId())
	return nil
}

func resourceDiskRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)
	getDiskResp, err := conn.SystemService().DisksService().
		DiskService(d.Id()).Get().Send()
	if err != nil {
		return err
	}

	disk, ok := getDiskResp.Disk()
	if !ok {
		d.SetId("")
		return nil
	}

	d.Set("name", disk.MustName())
	d.Set("size", disk.MustProvisionedSize())
	d.Set("format", disk.MustFormat())

	if sds, ok := disk.StorageDomains(); ok {
		if len(sds.Slice()) > 0 {
			d.Set("storage_domain_id", sds.Slice()[0].MustId())
		}
	}

	if shareable, ok := disk.Shareable(); ok {
		d.Set("shareable", shareable)
	}

	if sparse, ok := disk.Sparse(); ok {
		d.Set("sparse", sparse)
	}

	return nil
}

func resourceDiskDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	_, err := conn.SystemService().DisksService().
		DiskService(d.Id()).Remove().Send()
	if err != nil {
		return err
	}
	return nil
}
