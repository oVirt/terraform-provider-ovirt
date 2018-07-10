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

func resourceOvirtDisk() *schema.Resource {
	return &schema.Resource{
		Create: resourceOvirtDiskCreate,
		Read:   resourceOvirtDiskRead,
		Update: resourceOvirtDiskUpdate,
		Delete: resourceOvirtDiskDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
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

func resourceOvirtDiskCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	diskBuilder := ovirtsdk4.NewDiskBuilder().
		Name(d.Get("name").(string)).
		Format(ovirtsdk4.DiskFormat(d.Get("format").(string))).
		ProvisionedSize(int64(d.Get("size").(int))).
		StorageDomainsOfAny(
			ovirtsdk4.NewStorageDomainBuilder().
				Id(d.Get("storage_domain_id").(string)).
				MustBuild())
	if alias, ok := d.GetOk("alias"); ok {
		diskBuilder.Alias(alias.(string))
	}
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
	return resourceOvirtDiskRead(d, meta)
}

func resourceOvirtDiskUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)
	diskService := conn.SystemService().
		DisksService().
		DiskService(d.Id())

	diskGetResp, err := diskService.Get().
		Header("All-Content", "true").
		Send()
	if err != nil {
		return err
	}

	disk, ok := diskGetResp.Disk()
	if !ok {
		d.SetId("")
		return nil
	}
	vmSlice, ok := disk.Vms()
	// Disk has not yet attached to any VM
	if !ok || len(vmSlice.Slice()) == 0 {
		return fmt.Errorf("Only the disks attached to VMs can be updated")
	}

	attributeUpdate := false
	if d.HasChange("alias") && d.Get("alias").(string) != "" {
		disk.SetAlias(d.Get("alias").(string))
		attributeUpdate = true
	}

	if d.HasChange("size") {
		oldSizeValue, newSizeValue := d.GetChange("size")
		oldSize := oldSizeValue.(int)
		newSize := newSizeValue.(int)
		if oldSize > newSize {
			return fmt.Errorf("Only size extension is supported")
		}
		disk.SetProvisionedSize(int64(newSize))
		attributeUpdate = true
	}

	if attributeUpdate {
		// Only retrieve the first VM
		vmID := vmSlice.Slice()[0].MustId()
		attachmentService := conn.SystemService().
			VmsService().
			VmService(vmID).
			DiskAttachmentsService().
			AttachmentService(d.Id())
		getAttachResp, err := attachmentService.Get().Send()
		if err != nil {
			return nil
		}
		attachment, ok := getAttachResp.Attachment()
		if !ok {
			return nil
		}
		attachment.SetDisk(disk)
		// Call updating attachment
		_, err = attachmentService.Update().
			DiskAttachment(attachment).
			Send()
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceOvirtDiskRead(d *schema.ResourceData, meta interface{}) error {
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
	if alias, ok := disk.Alias(); ok {
		d.Set("alias", alias)
	}
	if shareable, ok := disk.Shareable(); ok {
		d.Set("shareable", shareable)
	}
	if sparse, ok := disk.Sparse(); ok {
		d.Set("sparse", sparse)
	}

	return nil
}

func resourceOvirtDiskDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	_, err := conn.SystemService().DisksService().
		DiskService(d.Id()).Remove().Send()
	if err != nil {
		return err
	}
	return nil
}
