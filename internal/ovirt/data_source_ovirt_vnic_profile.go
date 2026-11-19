package ovirt

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (p *provider) vnicProfileIdDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: p.vnicProfileDataSourceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "oVirt ID of the vnic profile.",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "oVirt name of the vnic profile.",
				Required:    true,
			},
		},
		Description: `Returns the ID of the given vnic_profile.`,
	}
}

func (p *provider) vnicProfileDataSourceRead(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	vnicProfileName := data.Get("name").(string)

	tpls, err := client.ListVNICProfiles()
	if err != nil {
		return errorToDiags("Listing all VNIC Profiles", err)
	}

	var id string
	for _, tpl := range tpls {
		if tpl.Name() == vnicProfileName {
			id = string(tpl.ID())
		}
	}

	data.SetId(id)

	return nil
}
