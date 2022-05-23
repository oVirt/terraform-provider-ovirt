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
	"os_type": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Description: "Operating system type.",
	},
	"placement_policy_affinity": {
		Type:             schema.TypeString,
		Optional:         true,
		RequiredWith:     []string{"placement_policy_host_ids"},
		Description:      "Affinity for placement policies. Must be one of: " + strings.Join(vmAffinityValues(), ", "),
		ValidateDiagFunc: validateEnum(vmAffinityValues()),
	},
	"placement_policy_host_ids": {
		Type:         schema.TypeSet,
		Optional:     true,
		RequiredWith: []string{"placement_policy_affinity"},
		Description:  "List of hosts to pin the VM to.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
}

func vmAffinityValues() []string {
	values := ovirtclient.VMAffinityValues()
	result := make([]string, len(values))
	for i, a := range values {
		result[i] = string(a)
	}
	return result
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
	var diags diag.Diagnostics
	for _, f := range []func(
		*schema.ResourceData,
		ovirtclient.BuildableVMParameters,
		diag.Diagnostics,
	) diag.Diagnostics{
		handleVMComment, handleVMCPUParameters, handleVMOSType, handleVMPlacementPolicy,
	} {
		diags = f(data, params, diags)
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

func handleVMPlacementPolicy(
	data *schema.ResourceData,
	params ovirtclient.BuildableVMParameters,
	diags diag.Diagnostics,
) diag.Diagnostics {
	placementPolicyBuilder := ovirtclient.NewVMPlacementPolicyParameters()
	hasPlacementPolicy := false
	var err error
	if a, ok := data.GetOk("placement_policy_affinity"); ok && a != "" {
		affinity := ovirtclient.VMAffinity(a.(string))
		if err := affinity.Validate(); err != nil {
			diags = append(diags, errorToDiag("create VM", err))
			return diags
		}
		placementPolicyBuilder, err = placementPolicyBuilder.WithAffinity(affinity)
		if err != nil {
			diags = append(diags, errorToDiag("add affinity to placement policy", err))
			return diags
		}
		hasPlacementPolicy = true
	}
	if hIDs, ok := data.GetOk("placement_policy_host_ids"); ok {
		hIDList := hIDs.(*schema.Set).List()
		hostIDs := make([]ovirtclient.HostID, len(hIDList))
		for i, hostID := range hIDList {
			hostIDs[i] = ovirtclient.HostID(hostID.(string))
		}
		placementPolicyBuilder, err = placementPolicyBuilder.WithHostIDs(hostIDs)
		if err != nil {
			diags = append(diags, errorToDiag("add host IDs to placement policy", err))
			return diags
		}
		hasPlacementPolicy = true
	}
	if hasPlacementPolicy {
		params.WithPlacementPolicy(placementPolicyBuilder)
	}
	return diags
}

func handleVMOSType(
	data *schema.ResourceData,
	params ovirtclient.BuildableVMParameters,
	diags diag.Diagnostics,
) diag.Diagnostics {
	if osType, ok := data.GetOk("os_type"); ok {
		osParams, err := ovirtclient.NewVMOSParameters().WithType(osType.(string))
		if err != nil {
			diags = append(diags, errorToDiag("add OS type to VM", err))
		}
		params.WithOS(osParams)
	}
	return diags
}

func handleVMCPUParameters(
	data *schema.ResourceData,
	params ovirtclient.BuildableVMParameters,
	diags diag.Diagnostics,
) diag.Diagnostics {
	if cpuCores, ok := data.GetOk("cpu_cores"); ok {
		cpuThreads := data.Get("cpu_threads").(int)
		cpuSockets := data.Get("cpu_sockets").(int)
		_, err := params.WithCPUParameters(uint(cpuCores.(int)), uint(cpuThreads), uint(cpuSockets))
		if err != nil {
			diags = append(diags, errorToDiag("add CPU parameters", err))
		}
	}
	return diags
}

func handleVMComment(
	data *schema.ResourceData,
	params ovirtclient.BuildableVMParameters,
	diags diag.Diagnostics,
) diag.Diagnostics {
	if comment, ok := data.GetOk("comment"); ok {
		_, err := params.WithComment(comment.(string))
		if err != nil {
			diags = append(
				diags,
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Invalid VM comment: %s", comment),
					Detail:   err.Error(),
				},
			)
		}
	}
	return diags
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
	if _, ok := data.GetOk("os_type"); ok || vm.OS().Type() != "other" {
		diags = setResourceField(data, "os_type", vm.OS().Type(), diags)
	}
	if pp, ok := vm.PlacementPolicy(); ok {
		diags = setResourceField(data, "placement_policy_host_ids", pp.HostIDs(), diags)
		diags = setResourceField(data, "placement_policy_affinity", pp.Affinity(), diags)
	}
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
