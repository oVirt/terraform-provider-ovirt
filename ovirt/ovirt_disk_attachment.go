package ovirt

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client"
)

var diskAttachmentSchema = map[string]*schema.Schema{
	"id": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"vm_id": {
		Type:             schema.TypeString,
		Required:         true,
		Description:      "ID of the VM the disk should be attached to.",
		ForceNew:         true,
		ValidateDiagFunc: validateUUID,
	},
	"disk_id": {
		Type:             schema.TypeString,
		Required:         true,
		Description:      "ID of the disk to attach. This disk must either be shared or not yet attached to a different VM.",
		ForceNew:         true,
		ValidateDiagFunc: validateUUID,
	},
	"disk_interface": {
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"Type of interface to use for attaching disk. One of: `%s`.",
			strings.Join(ovirtclient.DiskInterfaceValues().Strings(), "`, `"),
		),
		ForceNew:         true,
		ValidateDiagFunc: validateDiskInterface,
	},
}

func (p *provider) diskAttachmentResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: p.diskAttachmentCreate,
		ReadContext:   p.diskAttachmentRead,
		DeleteContext: p.diskAttachmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: p.diskAttachmentImport,
		},
		Schema: diskAttachmentSchema,
		Description: `The ovirt_disk_attachment resource attaches a single disk to a single VM. For controlling multiple attachments use ovirt_disk_attachments.

~> Do not use this resource when using ovirt_disk_attachments (plural) on the same VM as it will cause a ping-pong effect of resources being created and removed on each run.`,
	}
}

func (p *provider) diskAttachmentCreate(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	vmID := data.Get("vm_id").(string)
	diskID := data.Get("disk_id").(string)
	diskInterface := data.Get("disk_interface").(string)

	diskAttachment, err := client.CreateDiskAttachment(
		ovirtclient.VMID(vmID),
		ovirtclient.DiskID(diskID),
		ovirtclient.DiskInterface(diskInterface),
		ovirtclient.CreateDiskAttachmentParams(),
	)
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to create disk attachment.",
				Detail:   err.Error(),
			},
		}
	}

	return diskAttachmentResourceUpdate(diskAttachment, data)
}

func (p *provider) diskAttachmentRead(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	vmID := data.Get("vm_id").(string)
	attachment, err := client.GetDiskAttachment(
		ovirtclient.VMID(vmID),
		ovirtclient.DiskAttachmentID(data.Id()),
	)
	if isNotFound(err) {
		data.SetId("")
		return nil
	}
	return diskAttachmentResourceUpdate(attachment, data)
}

func (p *provider) diskAttachmentDelete(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	vmID := data.Get("vm_id").(string)
	if err := client.RemoveDiskAttachment(
		ovirtclient.VMID(vmID),
		ovirtclient.DiskAttachmentID(data.Id()),
	); err != nil {
		if isNotFound(err) {
			data.SetId("")
			return nil
		}
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to remove disk attachment.",
				Detail:   err.Error(),
			},
		}
	}
	data.SetId("")
	return nil
}

func (p *provider) diskAttachmentImport(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) ([]*schema.ResourceData, error) {
	client := p.client.WithContext(ctx)
	importID := data.Id()

	parts := strings.SplitN(importID, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf(
			"invalid import specification, the ID should be specified as: VMID/DiskAttachmentID",
		)
	}
	attachment, err := client.GetDiskAttachment(
		ovirtclient.VMID(parts[0]),
		ovirtclient.DiskAttachmentID(parts[1]),
	)
	if isNotFound(err) {
		return nil, fmt.Errorf("disk attachment with the specified VMID/ID %s not found (%w)", importID, err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to import disk_attachment %s (%w)", importID, err)
	}

	data.SetId(string(attachment.ID()))
	if err := data.Set("vm_id", attachment.VMID()); err != nil {
		return nil, fmt.Errorf("failed to set vm_id to %s", attachment.VMID())
	}
	if err := data.Set("disk_id", attachment.DiskID()); err != nil {
		return nil, fmt.Errorf("failed to set disk_id to %s", attachment.DiskID())
	}
	if err := data.Set("disk_interface", string(attachment.DiskInterface())); err != nil {
		return nil, fmt.Errorf("failed to set disk_interface to %s", attachment.DiskInterface())
	}
	return []*schema.ResourceData{data}, nil
}

func diskAttachmentResourceUpdate(disk ovirtclient.DiskAttachment, data *schema.ResourceData) diag.Diagnostics {
	data.SetId(string(disk.ID()))
	return nil
}
