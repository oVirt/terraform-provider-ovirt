// Copyright (C) 2017 Battelle Memorial Institute
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4"
)

func resourceOvirtDiskAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceOvirtDiskAttachmentCreate,
		Read:   resourceOvirtDiskAttachmentRead,
		Update: resourceOvirtDiskAttachmentUpdate,
		Delete: resourceOvirtDiskAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"vm_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"disk_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"active": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"bootable": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"interface": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"logical_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"pass_discard": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"read_only": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"use_scsi_reservation": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceOvirtDiskAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	diskID := d.Get("disk_id").(string)
	diskService := conn.SystemService().DisksService().DiskService(diskID)

	var disk *ovirtsdk4.Disk
	err := resource.Retry(30*time.Second, func() *resource.RetryError {
		getDiskResp, err := diskService.Get().Send()
		if err != nil {
			return resource.RetryableError(err)
		}
		disk = getDiskResp.MustDisk()

		if disk.MustStatus() == ovirtsdk4.DISKSTATUS_LOCKED {
			return resource.RetryableError(fmt.Errorf("disk is locked, wait for next check"))
		}
		return nil
	})
	if err != nil {
		return err
	}

	attachmentBuilder := ovirtsdk4.NewDiskAttachmentBuilder().
		Disk(disk).
		Interface(ovirtsdk4.DiskInterface(d.Get("interface").(string))).
		Bootable(d.Get("bootable").(bool)).
		Active(d.Get("active").(bool)).
		ReadOnly(d.Get("read_only").(bool)).
		UsesScsiReservation(d.Get("use_scsi_reservation").(bool))
	if logicalName, ok := d.GetOk("logical_name"); ok {
		attachmentBuilder.LogicalName(logicalName.(string))
	}
	if passDiscard, ok := d.GetOkExists("pass_discard"); ok {
		attachmentBuilder.PassDiscard(passDiscard.(bool))
	}
	attachment, err := attachmentBuilder.Build()
	if err != nil {
		return err
	}

	vmID := d.Get("vm_id").(string)
	addAttachmentResp, err := conn.SystemService().
		VmsService().
		VmService(vmID).
		DiskAttachmentsService().
		Add().
		Attachment(attachment).
		Send()
	if err != nil {
		return err
	}

	_, ok := addAttachmentResp.Attachment()
	if ok {
		d.SetId(vmID + ":" + diskID)
	}

	return resourceOvirtDiskAttachmentRead(d, meta)
}

func resourceOvirtDiskAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceOvirtDiskAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)
	// Disk ID is equals to its relevant Disk Attachment ID
	// Sess: https://github.com/oVirt/ovirt-engine/blob/68753f46f09419ddcdbb632453501273697d1a20/backend/manager/modules/restapi/types/src/main/java/org/ovirt/engine/api/restapi/types/DiskAttachmentMapper.java
	vmID, diskID, err := getVMIDAndDiskID(d, meta)
	if err != nil {
		return err
	}
	d.Set("vm_id", vmID)
	d.Set("disk_id", diskID)

	attachmentService := conn.SystemService().
		VmsService().
		VmService(vmID).
		DiskAttachmentsService().AttachmentService(diskID)
	attachmentResp, err := attachmentService.Get().Send()
	if err != nil {
		return err
	}
	attachment, ok := attachmentResp.Attachment()
	if !ok {
		d.SetId("")
		return nil
	}

	d.Set("active", attachment.MustActive())
	d.Set("bootable", attachment.MustBootable())
	d.Set("interface", attachment.MustInterface())
	d.Set("read_only", attachment.MustReadOnly())
	d.Set("use_scsi_reservation", attachment.MustUsesScsiReservation())
	if logicalName, ok := attachment.LogicalName(); ok {
		d.Set("logical_name", logicalName)
	}
	if passDiscard, ok := attachment.PassDiscard(); ok {
		d.Set("pass_discard", passDiscard)
	}

	return nil
}

func resourceOvirtDiskAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	vmID, diskID, err := getVMIDAndDiskID(d, meta)
	if err != nil {
		return err
	}

	vmService := conn.SystemService().VmsService().VmService(vmID)

	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		vmGetResp, err := vmService.Get().Send()
		if err != nil {
			return resource.RetryableError(err)
		}
		if vmGetResp.MustVm().MustStatus() != ovirtsdk4.VMSTATUS_DOWN {
			return resource.RetryableError(fmt.Errorf("The VM attached to is not down"))
		}
		return nil
	})

	_, err = vmService.
		DiskAttachmentsService().
		AttachmentService(diskID).
		Remove().
		Send()
	if err != nil {
		return err
	}

	return nil
}

func getVMIDAndDiskID(d *schema.ResourceData, meta interface{}) (string, string, error) {
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid resource id")
	}
	return parts[0], parts[1], nil
}
