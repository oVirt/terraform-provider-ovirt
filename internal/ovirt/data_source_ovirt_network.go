// Package ovirt provides datasource for managing oVirt logical networks
package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func (p *provider) networkDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: p.networkDataSourceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "oVirt ID of the network. Either name or id must be specified.",
				ValidateDiagFunc: validateUUID,
			},
			"name": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "Human-readable name of the logical network. Either name or id must be specified.",
				ValidateDiagFunc: validateNonEmpty,
			},
			"datacenter_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "ID of the data center the network belongs to. Used for disambiguation if multiple networks have the same name.",
				ValidateDiagFunc: validateUUID,
			},
		},
		Description: "The ovirt_network data source returns the ID and details of a logical network by its name or ID.",
	}
}

func (p *provider) networkDataSourceRead(
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
	var datacenterID string
	if dcVal, ok := data.GetOk("datacenter_id"); ok {
		datacenterID = dcVal.(string)
	}

	if id == "" && name == "" {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Missing network identifier",
				Detail:   "Either 'name' or 'id' must be specified to look up a network.",
			},
		}
	}

	var net ovirtclient.Network
	if id != "" {
		var err error
		net, err = client.GetNetwork(ovirtclient.NetworkID(id))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Failed to get network with ID %q", id),
					Detail:   err.Error(),
				},
			}
		}
		if name != "" && net.Name() != name {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Network name mismatch for ID %q", id),
					Detail:   fmt.Sprintf("Network with ID %q has name %q, but %q was specified.", id, net.Name(), name),
				},
			}
		}
		if datacenterID != "" && string(net.DatacenterID()) != datacenterID {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Network data center mismatch for ID %q", id),
					Detail:   fmt.Sprintf("Network with ID %q belongs to data center %q, but %q was specified.", id, net.DatacenterID(), datacenterID),
				},
			}
		}
	} else {
		networks, err := client.ListNetworks()
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to list networks",
					Detail:   err.Error(),
				},
			}
		}

		var matched []ovirtclient.Network
		for _, item := range networks {
			if item.Name() != name {
				continue
			}
			if datacenterID != "" && string(item.DatacenterID()) != datacenterID {
				continue
			}
			matched = append(matched, item)
		}

		if len(matched) == 0 {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Network with name %q not found", name),
					Detail:   fmt.Sprintf("No network found with name %q in the oVirt engine.", name),
				},
			}
		}

		if len(matched) > 1 {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Multiple networks found with name %q", name),
					Detail:   fmt.Sprintf("Found %d networks with name %q; specify datacenter_id to disambiguate.", len(matched), name),
				},
			}
		}

		net = matched[0]
	}

	data.SetId(string(net.ID()))

	diags := diag.Diagnostics{}
	diags = setResourceField(data, "name", net.Name(), diags)
	diags = setResourceField(data, "datacenter_id", string(net.DatacenterID()), diags)

	return diags
}
