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
		Type:             schema.TypeString,
		Required:         true,
		Description:      "User-provided name for the VM. Must only consist of lower- and uppercase letters, numbers, dash, underscore and dot.",
		ValidateDiagFunc: validateNonEmpty,
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
	"cpu_cores": {
		Type:             schema.TypeInt,
		Optional:         true,
		RequiredWith:     []string{"cpu_sockets", "cpu_threads"},
		Description:      "Number of CPU cores to allocate to the VM. If set, cpu_threads and cpu_sockets must also be specified.",
		ValidateDiagFunc: validatePositiveInt,
	},
	"cpu_threads": {
		Type:             schema.TypeInt,
		Optional:         true,
		RequiredWith:     []string{"cpu_sockets", "cpu_cores"},
		Description:      "Number of CPU threads to allocate to the VM. If set, cpu_cores and cpu_sockets must also be specified.",
		ValidateDiagFunc: validatePositiveInt,
	},
	"cpu_sockets": {
		Type:             schema.TypeInt,
		Optional:         true,
		RequiredWith:     []string{"cpu_threads", "cpu_cores"},
		Description:      "Number of CPU sockets to allocate to the VM. If set, cpu_cores and cpu_threads must also be specified.",
		ValidateDiagFunc: validatePositiveInt,
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
	client := p.client.WithContext(ctx)
	clusterID := data.Get("cluster_id").(string)
	templateID := data.Get("template_id").(string)

	params := ovirtclient.NewCreateVMParams()
	name := data.Get("name").(string)
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
	if cpuCores, ok := data.GetOk("cpu_cores"); ok {
		cpuThreads := data.Get("cpu_threads").(int)
		cpuSockets := data.Get("cpu_sockets").(int)
		_, err := params.WithCPUParameters(uint(cpuCores.(int)), uint(cpuThreads), uint(cpuSockets))
		if err != nil {
			return errorToDiags("add CPU parameters", err)
		}
	}

	vm, err := client.CreateVM(
		ovirtclient.ClusterID(clusterID),
		ovirtclient.TemplateID(templateID),
		name,
		params,
	)
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
	client := p.client.WithContext(ctx)
	id := data.Id()
	vm, err := client.GetVM(ovirtclient.VMID(id))
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
	data.SetId(string(vm.ID()))
	diags = setResourceField(data, "cluster_id", vm.ClusterID(), diags)
	diags = setResourceField(data, "template_id", vm.TemplateID(), diags)
	diags = setResourceField(data, "name", vm.Name(), diags)
	diags = setResourceField(data, "comment", vm.Comment(), diags)
	diags = setResourceField(data, "status", vm.Status(), diags)
	return diags
}

func (p *provider) vmDelete(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	if err := client.RemoveVM(ovirtclient.VMID(data.Id())); err != nil {
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
	client := p.client.WithContext(ctx)
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

	vm, err := client.UpdateVM(ovirtclient.VMID(data.Id()), params)
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
	client := p.client.WithContext(ctx)
	vm, err := client.GetVM(ovirtclient.VMID(data.Id()))
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
