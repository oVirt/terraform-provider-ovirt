// Package ovirt provides datasource for managing oVirt network interfaces
package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func (p *provider) nicDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: p.nicDataSourceRead,
		Schema: map[string]*schema.Schema{
			"vm_id": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "oVirt ID of the VM the NIC is attached to.",
				ValidateDiagFunc: validateUUID,
			},
			"id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "oVirt ID of the NIC. Either id or name must be specified.",
				ValidateDiagFunc: validateUUID,
			},
			"name": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "Name of the NIC. Either id or name must be specified.",
				ValidateDiagFunc: validateNonEmpty,
			},
			"vnic_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the VNIC profile associated with this NIC.",
			},
			"mac": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "MAC address of the NIC.",
			},
		},
		Description: "The ovirt_nic data source returns the details of a network interface attached to a VM by its ID or name.",
	}
}

func (p *provider) nicDataSourceRead(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	client := p.client.WithContext(ctx)

	vmID := data.Get("vm_id").(string)
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
				Summary:  "Missing NIC identifier",
				Detail:   "Either 'name' or 'id' must be specified along with 'vm_id' to look up a network interface.",
			},
		}
	}

	var nic ovirtclient.NIC
	if id != "" {
		var err error
		nic, err = client.GetNIC(ovirtclient.VMID(vmID), ovirtclient.NICID(id))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Failed to get NIC with ID %q on VM %q", id, vmID),
					Detail:   err.Error(),
				},
			}
		}
		if name != "" && nic.Name() != name {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("NIC name mismatch for ID %q", id),
					Detail:   fmt.Sprintf("NIC with ID %q has name %q, but %q was specified.", id, nic.Name(), name),
				},
			}
		}
	} else {
		nics, err := client.ListNICs(ovirtclient.VMID(vmID))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Failed to list NICs for VM %q", vmID),
					Detail:   err.Error(),
				},
			}
		}
		for _, candidate := range nics {
			if candidate.Name() == name {
				nic = candidate
				break
			}
		}
		if nic == nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("NIC with name %q not found", name),
					Detail:   fmt.Sprintf("No network interface with name %q was found on VM %q.", name, vmID),
				},
			}
		}
	}

	data.SetId(string(nic.ID()))

	diags := diag.Diagnostics{}
	diags = setResourceField(data, "vm_id", string(nic.VMID()), diags)
	diags = setResourceField(data, "name", nic.Name(), diags)
	diags = setResourceField(data, "vnic_profile_id", string(nic.VNICProfileID()), diags)
	diags = setResourceField(data, "mac", nic.Mac(), diags)

	return diags
}
