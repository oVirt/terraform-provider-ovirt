//nolint:revive
package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

var networkSchema = map[string]*schema.Schema{
	"id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "oVirt ID of this network.",
	},
	"name": {
		Type:             schema.TypeString,
		Required:         true,
		Description:      "User-provided name for the network.",
		ValidateDiagFunc: validateNonEmpty,
	},
	"datacenter_id": {
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		Description:      "Datacenter to create this network in.",
		ValidateDiagFunc: validateUUID,
	},
	"description": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "User-provided description for the network.",
	},
	"comment": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "User-provided comment for the network.",
	},
	"vlan": {
		Type:             schema.TypeInt,
		Optional:         true,
		Description:      "VLAN tag for the network (1-4094).",
		ValidateDiagFunc: validateVLAN,
	},
}

func validateVLAN(i interface{}, path cty.Path) diag.Diagnostics {
	v, ok := i.(int)
	if !ok {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid VLAN",
				Detail:        fmt.Sprintf("Expected integer for VLAN, got %T", i),
				AttributePath: path,
			},
		}
	}
	if v < 1 || v > 4094 {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid VLAN",
				Detail:        fmt.Sprintf("VLAN must be between 1 and 4094, got %d", v),
				AttributePath: path,
			},
		}
	}
	return nil
}

func (p *provider) networkResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: p.networkCreate,
		ReadContext:   p.networkRead,
		UpdateContext: p.networkUpdate,
		DeleteContext: p.networkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema:      networkSchema,
		Description: "The ovirt_network resource creates logical networks in oVirt.",
	}
}

func (p *provider) networkCreate(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	name := data.Get("name").(string)
	datacenterID := data.Get("datacenter_id").(string)

	params := ovirtclient.NewCreateNetworkParams()
	if desc, ok := data.GetOk("description"); ok {
		var err error
		params, err = params.WithDescription(desc.(string))
		if err != nil {
			return errorToDiags("set description", err)
		}
	}
	if comment, ok := data.GetOk("comment"); ok {
		var err error
		params, err = params.WithComment(comment.(string))
		if err != nil {
			return errorToDiags("set comment", err)
		}
	}
	if vlan, ok := data.GetOk("vlan"); ok {
		var err error
		params, err = params.WithVLAN(uint(vlan.(int)))
		if err != nil {
			return errorToDiags("set vlan", err)
		}
	}

	net, err := client.CreateNetwork(ovirtclient.DatacenterID(datacenterID), name, params)
	if err != nil {
		return errorToDiags(fmt.Sprintf("create network %s", name), err)
	}

	return networkToData(net, data)
}

func (p *provider) networkRead(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	net, err := client.GetNetwork(ovirtclient.NetworkID(data.Id()))
	if err != nil {
		if isNotFound(err) {
			data.SetId("")
			return nil
		}
		return errorToDiags(fmt.Sprintf("get network %s", data.Id()), err)
	}
	return networkToData(net, data)
}

func (p *provider) networkUpdate(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)

	params := ovirtclient.NewUpdateNetworkParams()
	if data.HasChange("name") {
		var err error
		params, err = params.WithName(data.Get("name").(string))
		if err != nil {
			return errorToDiags("set name", err)
		}
	}
	if data.HasChange("description") {
		var err error
		params, err = params.WithDescription(data.Get("description").(string))
		if err != nil {
			return errorToDiags("set description", err)
		}
	}
	if data.HasChange("comment") {
		var err error
		params, err = params.WithComment(data.Get("comment").(string))
		if err != nil {
			return errorToDiags("set comment", err)
		}
	}
	if data.HasChange("vlan") {
		if v, ok := data.GetOk("vlan"); ok {
			var err error
			params, err = params.WithVLAN(uint(v.(int)))
			if err != nil {
				return errorToDiags("set vlan", err)
			}
		}
	}

	net, err := client.UpdateNetwork(ovirtclient.NetworkID(data.Id()), params)
	if err != nil {
		return errorToDiags(fmt.Sprintf("update network %s", data.Id()), err)
	}
	return networkToData(net, data)
}

func (p *provider) networkDelete(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	if err := client.RemoveNetwork(ovirtclient.NetworkID(data.Id())); err != nil {
		if isNotFound(err) {
			data.SetId("")
			return nil
		}
		return errorToDiags(fmt.Sprintf("remove network %s", data.Id()), err)
	}
	data.SetId("")
	return nil
}

func networkToData(net ovirtclient.Network, data *schema.ResourceData) diag.Diagnostics {
	diags := diag.Diagnostics{}
	data.SetId(string(net.ID()))
	diags = appendDiags(diags, "set name on ovirt_network", data.Set("name", net.Name()))
	diags = appendDiags(diags, "set datacenter_id on ovirt_network", data.Set("datacenter_id", string(net.DatacenterID())))
	if desc := net.Description(); desc != nil {
		diags = appendDiags(diags, "set description on ovirt_network", data.Set("description", *desc))
	}
	if comment := net.Comment(); comment != nil {
		diags = appendDiags(diags, "set comment on ovirt_network", data.Set("comment", *comment))
	}
	if vlan := net.VLAN(); vlan != nil {
		diags = appendDiags(diags, "set vlan on ovirt_network", data.Set("vlan", int(*vlan)))
	}
	return diags
}
