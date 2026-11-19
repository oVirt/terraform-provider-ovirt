// Package ovirt provides datasource for managing oVirt data centers
package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func (p *provider) datacenterDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: p.datacenterDataSourceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "oVirt ID of the data center. Either name or id must be specified.",
				ValidateDiagFunc: validateUUID,
			},
			"name": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "Human-readable name of the data center. Either name or id must be specified.",
				ValidateDiagFunc: validateNonEmpty,
			},
		},
		Description: "The ovirt_datacenter data source returns the ID and details of a data center by its name or ID.",
	}
}

func (p *provider) datacenterDataSourceRead(
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
				Summary:  "Missing data center identifier",
				Detail:   "Either 'name' or 'id' must be specified to look up a data center.",
			},
		}
	}

	var dc ovirtclient.Datacenter
	if id != "" {
		var err error
		dc, err = client.GetDatacenter(ovirtclient.DatacenterID(id))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Failed to get data center with ID %q", id),
					Detail:   err.Error(),
				},
			}
		}
		if name != "" && dc.Name() != name {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Data center name mismatch for ID %q", id),
					Detail:   fmt.Sprintf("Data center with ID %q has name %q, but %q was specified.", id, dc.Name(), name),
				},
			}
		}
	} else {
		datacenters, err := client.ListDatacenters()
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to list data centers",
					Detail:   err.Error(),
				},
			}
		}

		var matched []ovirtclient.Datacenter
		for _, item := range datacenters {
			if item.Name() == name {
				matched = append(matched, item)
			}
		}

		if len(matched) == 0 {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Data center with name %q not found", name),
					Detail:   fmt.Sprintf("No data center found with name %q in the oVirt engine.", name),
				},
			}
		}

		if len(matched) > 1 {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Multiple data centers found with name %q", name),
					Detail:   fmt.Sprintf("Found %d data centers with name %q; data center names should be unique.", len(matched), name),
				},
			}
		}
		dc = matched[0]
	}

	data.SetId(string(dc.ID()))

	diags := diag.Diagnostics{}
	diags = setResourceField(data, "name", dc.Name(), diags)

	return diags
}
