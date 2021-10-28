package ovirt

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	ovirtclient "github.com/ovirt/go-ovirt-client"
)

var diskSchema = map[string]*schema.Schema{
	"storagedomain_id": {
		Type:             schema.TypeString,
		Required:         true,
		Description:      "ID of the storage domain to use for disk creation.",
		ForceNew:         true,
		ValidateFunc:     validateCompat(validateUUID),
	},
	"format": {
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"Format for the disk. One of: `%s`",
			strings.Join(ovirtclient.ImageFormatValues().Strings(), "`, `"),
		),
		ValidateFunc:     validateCompat(validateFormat),
		ForceNew:         true,
	},
	"size": {
		Type:             schema.TypeInt,
		Required:         true,
		Description:      "Disk size in bytes.",
		ValidateFunc:     validateCompat(validateDiskSize),
		ForceNew:         true,
	},
	"alias": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Human-readable alias for the disk.",
	},
	"sparse": {
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Description: "Use sparse provisioning for disk.",
	},
	"total_size": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Size of the actual image size on the disk in bytes.",
	},
	"status": {
		Type:     schema.TypeString,
		Computed: true,
		Description: fmt.Sprintf(
			"Status of the disk. One of: `%s`.",
			strings.Join(ovirtclient.VMStatusValues().Strings(), "`, `"),
		),
	},
}

func (p *provider) diskResource() *schema.Resource {
	return &schema.Resource{
		Create: crudCompat(p.diskCreate),
		Read:   crudCompat(p.diskRead),
		Update: crudCompat(p.diskUpdate),
		Delete: crudCompat(p.diskDelete),
		Importer: &schema.ResourceImporter{
			State: importCompat(p.diskImport),
		},
		Schema:      diskSchema,
		Description: "The ovirt_disk resource creates disks in oVirt.",
	}
}

func (p *provider) diskCreate(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	var err error

	storageDomainID := data.Get("storagedomain_id").(string)
	format := data.Get("format").(string)
	size := data.Get("size").(int)

	params := ovirtclient.CreateDiskParams()
	if alias, ok := data.GetOk("alias"); ok {
		params, err = params.WithAlias(alias.(string))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid alias value.",
					Detail:   err.Error(),
				},
			}
		}
	}
	if sparse, ok := data.GetOk("sparse"); ok {
		params, err = params.WithSparse(sparse.(bool))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid sparse value.",
					Detail:   err.Error(),
				},
			}
		}
	}

	disk, err := p.client.CreateDisk(
		storageDomainID,
		ovirtclient.ImageFormat(format),
		uint64(size),
		params,
		ovirtclient.ContextStrategy(ctx),
	)
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to create disk.",
				Detail:   err.Error(),
			},
		}
	}

	return diskResourceUpdate(disk, data)
}

func diskResourceUpdate(disk ovirtclient.Disk, data *schema.ResourceData) diag.Diagnostics {
	diags := diag.Diagnostics{}
	data.SetId(disk.ID())
	diags = setResourceField(data, "alias", disk.Alias(), diags)
	diags = setResourceField(data, "storagedomain_id", disk.StorageDomainID(), diags)
	diags = setResourceField(data, "format", string(disk.Format()), diags)
	diags = setResourceField(data, "size", disk.ProvisionedSize(), diags)
	diags = setResourceField(data, "sparse", disk.Sparse(), diags)
	diags = setResourceField(data, "total_size", disk.TotalSize(), diags)
	diags = setResourceField(data, "status", disk.Status(), diags)
	return diags
}

func (p *provider) diskRead(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	disk, err := p.client.GetDisk(data.Id(), ovirtclient.ContextStrategy(ctx))
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to fetch disk.",
				Detail:   err.Error(),
			},
		}
	}
	return diskResourceUpdate(disk, data)
}

func (p *provider) diskUpdate(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	params := ovirtclient.UpdateDiskParams()
	var err error
	if alias, ok := data.GetOk("alias"); ok {
		params, err = params.WithAlias(alias.(string))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid alias value.",
					Detail:   err.Error(),
				},
			}
		}
	}
	disk, err := p.client.UpdateDisk(data.Id(), params, ovirtclient.ContextStrategy(ctx))
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update disk.",
				Detail:   err.Error(),
			},
		}
	}
	return diskResourceUpdate(disk, data)
}

func (p *provider) diskDelete(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	if err := p.client.RemoveDisk(data.Id(), ovirtclient.ContextStrategy(ctx)); err != nil {
		if isNotFound(err) {
			data.SetId("")
			return nil
		}
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to remove disk.",
				Detail:   err.Error(),
			},
		}
	}
	data.SetId("")
	return nil
}

func (p *provider) diskImport(ctx context.Context, data *schema.ResourceData, _ interface{}) (
	[]*schema.ResourceData,
	error,
) {
	disk, err := p.client.GetDisk(data.Id(), ovirtclient.ContextStrategy(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to import disk %s (%w)", data.Id(), err)
	}
	diags := diskResourceUpdate(disk, data)
	if err := diagsToError(diags); err != nil {
		return nil, fmt.Errorf("failed to import disk (%w)", err)
	}
	return []*schema.ResourceData{
		data,
	}, nil
}
