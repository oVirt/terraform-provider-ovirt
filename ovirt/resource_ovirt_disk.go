// Copyright (C) 2017 Battelle Memorial Institute
// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
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
				ForceNew: false,
			},
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"format": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(ovirtsdk4.DISKFORMAT_COW),
					string(ovirtsdk4.DISKFORMAT_RAW),
				}, false),
			},
			"storage_domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: false,
			},
			"shareable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"sparse": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			// "qcow_version" is the only field supporting Disk-Update
			// See: http://ovirt.github.io/ovirt-engine-api-model/4.3/#services/disk/methods/update
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
	diskID := addResp.MustDisk().MustId()
	d.SetId(diskID)

	// Wait for disk is ready
	err = conn.WaitForDisk(diskID, ovirtsdk4.DISKSTATUS_OK, 2*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to wait for disk status to be OK: %s", err)
	}

	return resourceOvirtDiskRead(d, meta)
}

func resourceOvirtDiskUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	vml, err := getAttachedVMsOfDisk(d.Id(), meta)
	if err != nil {
		if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
			d.SetId("")
			return nil
		}
	}
	// Disk has not yet attached to any VM
	if len(vml) == 0 {
		return fmt.Errorf("Only the disks attached to VMs can be updated")
	}

	paramDisk := ovirtsdk4.NewDiskBuilder()

	attributeUpdate := false
	if d.HasChange("name") && d.Get("name").(string) != "" {
		paramDisk.Name(d.Get("name").(string))
		attributeUpdate = true
	}
	if d.HasChange("alias") && d.Get("alias").(string) != "" {
		paramDisk.Alias(d.Get("alias").(string))
		attributeUpdate = true
	}
	if d.HasChange("size") {
		oldSizeValue, newSizeValue := d.GetChange("size")
		oldSize := oldSizeValue.(int)
		newSize := newSizeValue.(int)
		if oldSize > newSize {
			return fmt.Errorf("Only size extension is supported")
		}
		paramDisk.ProvisionedSize(int64(newSize))
		attributeUpdate = true
	}

	if attributeUpdate {
		// Only retrieve the first VM
		vmID := vml[0].MustId()
		attachmentService := conn.SystemService().
			VmsService().
			VmService(vmID).
			DiskAttachmentsService().
			AttachmentService(d.Id())

		_, err := attachmentService.Update().DiskAttachment(
			ovirtsdk4.NewDiskAttachmentBuilder().
				Disk(
					paramDisk.MustBuild()).
				MustBuild()).
			Send()
		if err != nil {
			return err
		}

		err = conn.WaitForDisk(d.Id(), ovirtsdk4.DISKSTATUS_OK, 1*time.Minute)
		if err != nil {
			return err
		}
	}

	return resourceOvirtDiskRead(d, meta)
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

	rmRequest := conn.SystemService().DisksService().
		DiskService(d.Id()).Remove()

	// Find VMs attached
	vml, err := getAttachedVMsOfDisk(d.Id(), meta)
	if err != nil {
		if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
			return nil
		}
		return err
	}
	// Shutdown VMs attached
	if len(vml) > 0 {
		for _, v := range vml {
			err := tryShutdownVM(v.MustId(), meta)
			if err != nil {
				return fmt.Errorf("[DEBUG] Failed to shutdown VM (%s) attached: %s", v.MustId(), err)
			}
		}
	}

	err = resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, e := rmRequest.Send()
		if e != nil {
			if _, ok := e.(*ovirtsdk4.NotFoundError); ok {
				return nil
			}
			return resource.RetryableError(fmt.Errorf("failed to delete disk: %s, wait for next check", e))
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func getAttachedVMsOfDisk(diskID string, meta interface{}) ([]*ovirtsdk4.Vm, error) {
	conn := meta.(*ovirtsdk4.Connection)

	diskService := conn.SystemService().DisksService().DiskService(diskID)
	getDiskResp, err := diskService.Get().
		Header("All-Content", "true").
		Send()
	if err != nil {
		return nil, err
	}

	if disk, ok := getDiskResp.Disk(); ok {
		if vmSlice, ok := disk.Vms(); ok {
			return vmSlice.Slice(), nil
		}
	}
	return nil, nil
}
