package ovirt

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func (p *provider) vnicProfileDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: p.vnicProfileDataSourceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "oVirt ID of the VNIC profile.",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Human-readable name of the VNIC profile.",
				Required:    true,
			},
			"network_id": {
				Type:        schema.TypeString,
				Description: "ID of the logical network the VNIC profile belongs to. If multiple profiles have the same name, this disambiguates them.",
				Optional:    true,
				Computed:    true,
			},
		},
		Description: `The ovirt_vnic_profile data source returns the ID and details of a VNIC profile by its name, with optional filtering by network_id.`,
	}
}

func (p *provider) vnicProfileDataSourceRead(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	vnicProfileName := data.Get("name").(string)
	networkIDFilter, hasNetworkID := data.GetOk("network_id")

	profiles, err := client.ListVNICProfiles()
	if err != nil {
		return errorToDiags("Listing all VNIC Profiles", err)
	}

	var matches []ovirtclient.VNICProfile
	for _, profile := range profiles {
		if profile.Name() == vnicProfileName {
			if hasNetworkID && string(profile.NetworkID()) != networkIDFilter.(string) {
				continue
			}
			matches = append(matches, profile)
		}
	}

	if len(matches) == 0 {
		if hasNetworkID {
			return diag.Errorf("No VNIC profile found with name %q and network_id %q", vnicProfileName, networkIDFilter.(string))
		}
		return diag.Errorf("No VNIC profile found with name %q", vnicProfileName)
	}

	if len(matches) > 1 {
		return diag.Errorf("Multiple VNIC profiles (%d) found with name %q. Please specify network_id to disambiguate.", len(matches), vnicProfileName)
	}

	selected := matches[0]
	data.SetId(string(selected.ID()))
	if err := data.Set("name", selected.Name()); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("network_id", string(selected.NetworkID())); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
