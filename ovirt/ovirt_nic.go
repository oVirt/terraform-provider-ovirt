package ovirt

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client"
)

var nicSchema = map[string]*schema.Schema{
	"id": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"vnic_profile_id": {
		Type:             schema.TypeString,
		Required:         true,
		Description:      "ID of the VNIC profile to associate with this NIC.",
		ValidateDiagFunc: validateUUID,
	},
	"vm_id": {
		Type:             schema.TypeString,
		Required:         true,
		Description:      "ID of the VM to attach this NIC to.",
		ForceNew:         true,
		ValidateDiagFunc: validateUUID,
	},
	"name": {
		Type:             schema.TypeString,
		Required:         true,
		Description:      "Human-readable name for the NIC.",
		ValidateDiagFunc: validateNonEmpty,
	},
}

func (p *provider) nicResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: p.nicCreate,
		ReadContext:   p.nicRead,
		UpdateContext: p.nicUpdate,
		DeleteContext: p.nicDelete,
		Importer: &schema.ResourceImporter{
			StateContext: p.nicImport,
		},
		Schema:      nicSchema,
		Description: "The ovirt_nic resource creates network interfaces in oVirt.",
	}
}

func (p *provider) nicCreate(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	vmID := data.Get("vm_id").(string)
	vnicProfileID := data.Get("vnic_profile_id").(string)
	name := data.Get("name").(string)

	nic, err := p.client.CreateNIC(vmID, vnicProfileID, name, nil, ovirtclient.ContextStrategy(ctx))
	if err != nil {
		return errorToDiags("create NIC", err)
	}

	return nicResourceUpdate(nic, data)
}

func (p *provider) nicRead(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	id := data.Id()
	vmID := data.Get("vm_id").(string)
	nic, err := p.client.GetNIC(vmID, id, ovirtclient.ContextStrategy(ctx))
	if err != nil {
		if isNotFound(err) {
			data.SetId("")
			return nil
		}
		return errorToDiags("get NIC", err)
	}
	return nicResourceUpdate(nic, data)
}

func (p *provider) nicUpdate(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	params := ovirtclient.UpdateNICParams()
	var err error
	if vnicProfileID, ok := data.GetOk("vnic_profile_id"); ok {
		params, err = params.WithVNICProfileID(vnicProfileID.(string))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid vnic_profile_id value.",
					Detail:   err.Error(),
				},
			}
		}
	}
	if name, ok := data.GetOk("name"); ok {
		params, err = params.WithName(name.(string))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid name value.",
					Detail:   err.Error(),
				},
			}
		}
	}
	vmID, ok := data.GetOk("vm_id")
	if !ok {
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid name value.",
					Detail:   err.Error(),
				},
			}
		}
	}
	nic, err := p.client.UpdateNIC(vmID.(string), data.Id(), params, ovirtclient.ContextStrategy(ctx))
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update nic.",
				Detail:   err.Error(),
			},
		}
	}
	return nicResourceUpdate(nic, data)
}

func (p *provider) nicDelete(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	id := data.Id()
	vmID := data.Get("vm_id").(string)
	if err := p.client.RemoveNIC(vmID, id, ovirtclient.ContextStrategy(ctx)); err != nil {
		if !isNotFound(err) {
			return errorToDiags("get NIC", err)
		}
	}
	data.SetId("")
	return nil
}

func (p *provider) nicImport(ctx context.Context, data *schema.ResourceData, _ interface{}) (
	[]*schema.ResourceData,
	error,
) {
	importID := data.Id()

	parts := strings.SplitN(importID, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf(
			"invalid import specification, the ID should be specified as: VMID/NICID",
		)
	}
	nic, err := p.client.GetNIC(parts[0], parts[1], ovirtclient.ContextStrategy(ctx))
	if err != nil {
		return nil, err
	}
	if diags := nicResourceUpdate(nic, data); diags.HasError() {
		return nil, diagsToError(diags)
	}
	return []*schema.ResourceData{data}, nil
}

func nicResourceUpdate(nic ovirtclient.NIC, data *schema.ResourceData) diag.Diagnostics {
	diags := diag.Diagnostics{}
	data.SetId(nic.ID())
	diags = setResourceField(data, "vnic_profile_id", nic.VNICProfileID(), diags)
	diags = setResourceField(data, "name", nic.Name(), diags)
	diags = setResourceField(data, "vm_id", nic.VMID(), diags)
	return diags
}
