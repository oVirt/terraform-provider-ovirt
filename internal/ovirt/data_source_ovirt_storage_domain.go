// Package ovirt provides datasource for managing oVirt storage domain
package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func (p *provider) storageDomainDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: p.storageDomainDataSourceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "oVirt ID of the storage domain. Either name or id must be specified.",
				ValidateDiagFunc: validateUUID,
			},
			"name": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "Human-readable name of the storage domain. Either name or id must be specified.",
				ValidateDiagFunc: validateNonEmpty,
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the storage domain (e.g. active, etc.).",
			},
			"available": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Available space on the storage domain in bytes.",
			},
			"storage_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the storage domain (e.g. nfs, iscsi, fcp, etc.).",
			},
			"external_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "External status of the storage domain.",
			},
		},
		Description: "The ovirt_storage_domain data source returns the ID and details of a storage domain by its name or ID.",
	}
}

func (p *provider) storageDomainDataSourceRead(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	client := p.client.WithContext(ctx)

	var id string
	if idVal, ok := data.GetOk("id"); ok {
		id = idVal.(string)
	}
	var name string
	if nameVal, ok := data.GetOk("name"); ok {
		name = nameVal.(string)
	}

	if id == "" && name == "" {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Missing storage domain identifier",
				Detail:   "Either 'name' or 'id' must be specified to look up a storage domain.",
			},
		}
	}

	var sd ovirtclient.StorageDomain
	if id != "" {
		var err error
		sd, err = client.GetStorageDomain(ovirtclient.StorageDomainID(id))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Failed to get storage domain with ID %q", id),
					Detail:   err.Error(),
				},
			}
		}
		if name != "" && sd.Name() != name {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Storage domain name mismatch for ID %q", id),
					Detail:   fmt.Sprintf("Storage domain with ID %q has name %q, but %q was specified.", id, sd.Name(), name),
				},
			}
		}
	} else {
		storageDomains, err := client.ListStorageDomains()
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to list storage domains",
					Detail:   err.Error(),
				},
			}
		}

		var matched []ovirtclient.StorageDomain
		for _, item := range storageDomains {
			if item.Name() == name {
				matched = append(matched, item)
			}
		}

		if len(matched) == 0 {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Storage domain with name %q not found", name),
					Detail:   fmt.Sprintf("No storage domain found with name %q in the oVirt engine.", name),
				},
			}
		}
		sd = matched[0]
	}

	data.SetId(string(sd.ID()))

	diags := diag.Diagnostics{}
	diags = setResourceField(data, "name", sd.Name(), diags)
	diags = setResourceField(data, "status", string(sd.Status()), diags)
	diags = setResourceField(data, "available", sd.Available(), diags)
	diags = setResourceField(data, "storage_type", string(sd.StorageType()), diags)
	diags = setResourceField(data, "external_status", string(sd.ExternalStatus()), diags)

	return diags
}
