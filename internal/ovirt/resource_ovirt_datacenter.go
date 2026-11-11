//nolint:revive
package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

var datacenterSchema = map[string]*schema.Schema{
	"id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "oVirt ID of this datacenter.",
	},
	"name": {
		Type:             schema.TypeString,
		Required:         true,
		Description:      "User-provided name for the datacenter.",
		ValidateDiagFunc: validateNonEmpty,
	},
	"description": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "User-provided description for the datacenter.",
	},
	"comment": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "User-provided comment for the datacenter.",
	},
	"local": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		ForceNew:    true,
		Description: "Whether the datacenter uses local storage (true) or shared storage (false). Defaults to false.",
	},
	"quota_mode": {
		Type:             schema.TypeString,
		Optional:         true,
		Computed:         true,
		ValidateDiagFunc: validateEnum([]string{"disabled", "audit", "enabled"}),
		Description:      "Quota enforcement mode for the datacenter. Valid values: disabled, audit, enabled.",
	},
}

func (p *provider) datacenterResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: p.datacenterCreate,
		ReadContext:   p.datacenterRead,
		UpdateContext: p.datacenterUpdate,
		DeleteContext: p.datacenterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema:      datacenterSchema,
		Description: "The ovirt_datacenter resource creates datacenters in oVirt.",
	}
}

func (p *provider) datacenterCreate(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	name := data.Get("name").(string)

	params := ovirtclient.NewCreateDatacenterParams()
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
	if local, ok := data.GetOkExists("local"); ok {
		var err error
		params, err = params.WithLocal(local.(bool))
		if err != nil {
			return errorToDiags("set local", err)
		}
	}
	if qm, ok := data.GetOk("quota_mode"); ok {
		var err error
		params, err = params.WithQuotaMode(ovirtclient.QuotaMode(qm.(string)))
		if err != nil {
			return errorToDiags("set quota_mode", err)
		}
	}

	dc, err := client.CreateDatacenter(name, params)
	if err != nil {
		return errorToDiags(fmt.Sprintf("create datacenter %s", name), err)
	}

	return datacenterToData(dc, data)
}

func (p *provider) datacenterRead(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	dc, err := client.GetDatacenter(ovirtclient.DatacenterID(data.Id()))
	if err != nil {
		if isNotFound(err) {
			data.SetId("")
			return nil
		}
		return errorToDiags(fmt.Sprintf("get datacenter %s", data.Id()), err)
	}
	return datacenterToData(dc, data)
}

func (p *provider) datacenterUpdate(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)

	params := ovirtclient.NewUpdateDatacenterParams()
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
	if data.HasChange("quota_mode") {
		if qm, ok := data.GetOk("quota_mode"); ok {
			var err error
			params, err = params.WithQuotaMode(ovirtclient.QuotaMode(qm.(string)))
			if err != nil {
				return errorToDiags("set quota_mode", err)
			}
		}
	}

	dc, err := client.UpdateDatacenter(ovirtclient.DatacenterID(data.Id()), params)
	if err != nil {
		return errorToDiags(fmt.Sprintf("update datacenter %s", data.Id()), err)
	}
	return datacenterToData(dc, data)
}

func (p *provider) datacenterDelete(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	if err := client.RemoveDatacenter(ovirtclient.DatacenterID(data.Id())); err != nil {
		if isNotFound(err) {
			data.SetId("")
			return nil
		}
		return errorToDiags(fmt.Sprintf("remove datacenter %s", data.Id()), err)
	}
	data.SetId("")
	return nil
}

func datacenterToData(dc ovirtclient.Datacenter, data *schema.ResourceData) diag.Diagnostics {
	diags := diag.Diagnostics{}
	data.SetId(string(dc.ID()))
	diags = appendDiags(diags, "set name on ovirt_datacenter", data.Set("name", dc.Name()))
	diags = appendDiags(diags, "set local on ovirt_datacenter", data.Set("local", dc.Local()))
	if desc := dc.Description(); desc != nil {
		diags = appendDiags(diags, "set description on ovirt_datacenter", data.Set("description", *desc))
	} else {
		diags = appendDiags(diags, "clear description on ovirt_datacenter", data.Set("description", nil))
	}
	if comment := dc.Comment(); comment != nil {
		diags = appendDiags(diags, "set comment on ovirt_datacenter", data.Set("comment", *comment))
	} else {
		diags = appendDiags(diags, "clear comment on ovirt_datacenter", data.Set("comment", nil))
	}
	if qm := dc.QuotaMode(); qm != nil {
		diags = appendDiags(diags, "set quota_mode on ovirt_datacenter", data.Set("quota_mode", string(*qm)))
	}
	return diags
}
