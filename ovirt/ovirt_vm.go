package ovirt

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client"
)

var vmSchema = map[string]*schema.Schema{
	"id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "oVirt ID of this VM.",
	},
	"name": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "User-provided name for the VM. Must only consist of lower- and uppercase letters, numbers, dash, underscore and dot.",
	},
	"comment": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "User-provided comment for the VM.",
	},
	"cluster_id": {
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		Description:      "Cluster to create this VM on.",
		ValidateDiagFunc: validateUUID,
	},
	"template_id": {
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		Description:      "Base template for this VM.",
		ValidateDiagFunc: validateUUID,
	},
	"status": {
		Type:     schema.TypeString,
		Computed: true,
		Description: fmt.Sprintf(
			"Status of the virtual machine. One of: `%s`.",
			strings.Join(ovirtclient.VMStatusValues().Strings(), "`, `"),
		),
	},
}

func (p *provider) vmResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: p.vmCreate,
		ReadContext:   p.vmRead,
		UpdateContext: p.vmUpdate,
		DeleteContext: p.vmDelete,
		Importer: &schema.ResourceImporter{
			StateContext: p.vmImport,
		},
		Schema:      vmSchema,
		Description: "The ovirt_vm resource creates a virtual machine in oVirt.",
	}
}

func (p *provider) vmCreate(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	clusterID := data.Get("cluster_id").(string)
	templateID := data.Get("template_id").(string)

	params := ovirtclient.CreateVMParams()
	if name, ok := data.GetOk("name"); ok {
		_, err := params.WithName(name.(string))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Invalid VM name: %s", name),
					Detail:   err.Error(),
				},
			}
		}
	}
	if comment, ok := data.GetOk("comment"); ok {
		_, err := params.WithComment(comment.(string))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Invalid VM comment: %s", comment),
					Detail:   err.Error(),
				},
			}
		}
	}

	vm, err := p.client.CreateVM(clusterID, templateID, params, ovirtclient.ContextStrategy(ctx))
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to create VM",
				Detail:   err.Error(),
			},
		}
	}

	return vmResourceUpdate(vm, data)
}

func (p *provider) vmRead(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	id := data.Id()
	vm, err := p.client.GetVM(id, ovirtclient.ContextStrategy(ctx))
	if err != nil {
		if isNotFound(err) {
			data.SetId("")
			return nil
		}
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to fetch VM %s", id),
				Detail:   err.Error(),
			},
		}
	}
	return vmResourceUpdate(vm, data)
}

// vmResourceUpdate takes the VM object and converts it into Terraform resource data.
func vmResourceUpdate(vm ovirtclient.VMData, data *schema.ResourceData) diag.Diagnostics {
	diags := diag.Diagnostics{}
	data.SetId(vm.ID())
	diags = setResourceField(data, "cluster_id", vm.ClusterID(), diags)
	diags = setResourceField(data, "template_id", vm.TemplateID(), diags)
	diags = setResourceField(data, "name", vm.Name(), diags)
	diags = setResourceField(data, "comment", vm.Comment(), diags)
	diags = setResourceField(data, "status", vm.Status(), diags)
	return diags
}

func (p *provider) vmDelete(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	if err := p.client.RemoveVM(data.Id(), ovirtclient.ContextStrategy(ctx)); err != nil {
		if isNotFound(err) {
			data.SetId("")
			return nil
		}
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       fmt.Sprintf("Failed to remove VM %s", data.Id()),
				Detail:        err.Error(),
				AttributePath: nil,
			},
		}
	}
	data.SetId("")
	return nil
}

func (p *provider) vmUpdate(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	params := ovirtclient.UpdateVMParams()
	if name, ok := data.GetOk("name"); ok {
		_, err := params.WithName(name.(string))
		if err != nil {
			diags = append(
				diags,
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid VM name",
					Detail:   err.Error(),
				},
			)
		}
	}
	if name, ok := data.GetOk("comment"); ok {
		_, err := params.WithComment(name.(string))
		if err != nil {
			diags = append(
				diags,
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid VM comment",
					Detail:   err.Error(),
				},
			)
		}
	}
	if len(diags) > 0 {
		return diags
	}

	vm, err := p.client.UpdateVM(data.Id(), params, ovirtclient.ContextStrategy(ctx))
	if isNotFound(err) {
		data.SetId("")
	}
	if err != nil {
		diags = append(
			diags,
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to update VM %s", data.Id()),
				Detail:   err.Error(),
			},
		)
		return diags
	}
	return vmResourceUpdate(vm, data)
}

func (p *provider) vmImport(ctx context.Context, data *schema.ResourceData, _ interface{}) (
	[]*schema.ResourceData,
	error,
) {
	vm, err := p.client.GetVM(data.Id(), ovirtclient.ContextStrategy(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to import VM %s (%w)", data.Id(), err)
	}
	d := vmResourceUpdate(vm, data)
	if err := diagsToError(d); err != nil {
		return nil, fmt.Errorf("failed to import VM %s (%w)", data.Id(), err)
	}
	return []*schema.ResourceData{
		data,
	}, nil
}
