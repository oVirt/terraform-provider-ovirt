// Package ovirt provides datasource for listing oVirt network interfaces
package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func (p *provider) nicsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: p.nicsDataSourceRead,
		Schema: map[string]*schema.Schema{
			"vm_id": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "oVirt ID of the VM to list the network interfaces for.",
				ValidateDiagFunc: validateUUID,
			},
			"nics": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of network interfaces attached to the VM.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the network interface.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the network interface.",
						},
						"vnic_profile_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the VNIC profile associated with this NIC.",
						},
						"mac": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "MAC address of the network interface.",
						},
					},
				},
			},
		},
		Description: "The ovirt_nics data source returns a set of all network interfaces attached to a virtual machine.",
	}
}

func (p *provider) nicsDataSourceRead(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	client := p.client.WithContext(ctx)

	vmID := data.Get("vm_id").(string)
	nicsList, err := client.ListNICs(ovirtclient.VMID(vmID))
	if err != nil {
		return errorToDiags(fmt.Sprintf("list NICs of VM %s", vmID), err)
	}

	nics := make([]map[string]interface{}, 0, len(nicsList))
	for _, nic := range nicsList {
		entry := map[string]interface{}{
			"id":              string(nic.ID()),
			"name":            nic.Name(),
			"vnic_profile_id": string(nic.VNICProfileID()),
			"mac":             nic.Mac(),
		}
		nics = append(nics, entry)
	}

	if err := data.Set("nics", nics); err != nil {
		return errorToDiags("set nics", err)
	}

	data.SetId(vmID)

	return nil
}
