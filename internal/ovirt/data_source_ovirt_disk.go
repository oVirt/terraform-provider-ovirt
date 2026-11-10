// Package ovirt provides datasource for managing oVirt disks
package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func (p *provider) diskDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: p.diskDataSourceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "oVirt ID of the disk. Either alias or id must be specified.",
				ValidateDiagFunc: validateUUID,
			},
			"alias": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "Human-readable alias of the disk. Either alias or id must be specified.",
				ValidateDiagFunc: validateNonEmpty,
			},
			"storage_domain_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "ID of the storage domain the disk belongs to. Used to filter/disambiguate disks with the same alias.",
				ValidateDiagFunc: validateUUID,
			},
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Disk provisioned size in bytes.",
			},
			"total_size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Size of the actual image on the disk in bytes.",
			},
			"format": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Format of the disk image (e.g. raw, cow).",
			},
			"sparse": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether sparse provisioning is enabled for the disk.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the disk (e.g. ok, locked, etc.).",
			},
		},
		Description: "The ovirt_disk data source returns the ID and details of a disk by its alias or ID.",
	}
}

func (p *provider) diskDataSourceRead(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	client := p.client.WithContext(ctx)

	var id string
	if idVal, ok := data.GetOk("id"); ok {
		id = idVal.(string)
	}
	var alias string
	if aliasVal, ok := data.GetOk("alias"); ok {
		alias = aliasVal.(string)
	}
	var storageDomainID string
	if sdVal, ok := data.GetOk("storage_domain_id"); ok {
		storageDomainID = sdVal.(string)
	}

	if id == "" && alias == "" {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Missing disk identifier",
				Detail:   "Either 'alias' or 'id' must be specified to look up a disk.",
			},
		}
	}

	var disk ovirtclient.Disk
	if id != "" {
		var err error
		disk, err = client.GetDisk(ovirtclient.DiskID(id))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Failed to get disk with ID %q", id),
					Detail:   err.Error(),
				},
			}
		}
		if alias != "" && disk.Alias() != alias {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Disk alias mismatch for ID %q", id),
					Detail:   fmt.Sprintf("Disk with ID %q has alias %q, but %q was specified.", id, disk.Alias(), alias),
				},
			}
		}
	} else {
		disks, err := client.ListDisks()
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to list disks",
					Detail:   err.Error(),
				},
			}
		}

		var matched []ovirtclient.Disk
		for _, item := range disks {
			if item.Alias() != alias {
				continue
			}
			if storageDomainID != "" {
				sdFound := false
				for _, sdID := range item.StorageDomainIDs() {
					if string(sdID) == storageDomainID {
						sdFound = true
						break
					}
				}
				if !sdFound {
					continue
				}
			}
			matched = append(matched, item)
		}

		if len(matched) == 0 {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Disk with alias %q not found", alias),
					Detail:   fmt.Sprintf("No disk found with alias %q in the oVirt engine.", alias),
				},
			}
		}

		if len(matched) > 1 {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Multiple disks found with alias %q", alias),
					Detail:   fmt.Sprintf("Found %d disks with alias %q; specify storage_domain_id to disambiguate.", len(matched), alias),
				},
			}
		}
		disk = matched[0]
	}

	data.SetId(string(disk.ID()))

	diags := diag.Diagnostics{}
	diags = setResourceField(data, "alias", disk.Alias(), diags)
	diags = setResourceField(data, "format", string(disk.Format()), diags)
	diags = setResourceField(data, "size", disk.ProvisionedSize(), diags)
	diags = setResourceField(data, "total_size", disk.TotalSize(), diags)
	diags = setResourceField(data, "sparse", disk.Sparse(), diags)
	diags = setResourceField(data, "status", string(disk.Status()), diags)
	if len(disk.StorageDomainIDs()) > 0 {
		diags = setResourceField(data, "storage_domain_id", string(disk.StorageDomainIDs()[0]), diags)
	}

	return diags
}
