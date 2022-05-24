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
	"cpu_mode": {
		Type:     schema.TypeString,
		Optional: true,
		Description: fmt.Sprintf(
			"Sets the CPU mode for the VM. Can be one of: %s",
			strings.Join(cpuModeValues(), ", "),
		),
		ValidateDiagFunc: validateEnum(cpuModeValues()),
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
	"template_disk_attachment_override": {
		Type:        schema.TypeSet,
		Optional:    true,
		ForceNew:    true,
		Description: "Override parameters for disks obtained from templates.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"disk_id": {
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					Description:      "ID of the disk to be changed.",
					ValidateDiagFunc: validateUUID,
				},
				"format": {
					Type:             schema.TypeString,
					Optional:         true,
					ForceNew:         true,
					Description:      "Disk format for the override. Can be 'raw' or 'cow'.",
					ValidateDiagFunc: validateFormat,
				},
				"sparse": {
					Type:        schema.TypeBool,
					Optional:    true,
					ForceNew:    true,
					Description: "Sparse-provision the disk.",
				},
			},
		},
	},
	"initialization_custom_script": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Custom script that passed to VM during initialization.",
	},
	"initialization_hostname": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "hostname that is set during initialization.",
	},
	"memory": {
		Type:             schema.TypeInt,
		Optional:         true,
		Description:      "Memory to assign to the VM in bytes.",
		ValidateDiagFunc: validatePositiveInt,
		RequiredWith:     []string{"maximum_memory"},
	},
	"maximum_memory": {
		Type:             schema.TypeInt,
		Optional:         true,
		Description:      "Maximum memory to assign to the VM in the memory policy in bytes.",
		ValidateDiagFunc: validatePositiveInt,
	},
	"memory_ballooning": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Turn memory ballooning on or off for the VM.",
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
		handleVMComment,
		handleVMCPUParameters,
		handleVMOSType,
		handleVMPlacementPolicy,
		handleVMInitialization,
		handleTemplateDiskAttachmentOverride,
		handleVMMemory,
		handleVMMemoryPolicy,
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

func handleTemplateDiskAttachmentOverride(
	data *schema.ResourceData,
	params ovirtclient.BuildableVMParameters,
	diags diag.Diagnostics,
) diag.Diagnostics {
	templateDiskAttachments, ok := data.GetOk("template_disk_attachment_override")
	if !ok {
		return diags
	}
	templateDiskAttachmentsSet := templateDiskAttachments.(*schema.Set)
	disks := make([]ovirtclient.OptionalVMDiskParameters, len(templateDiskAttachmentsSet.List()))
	for i, item := range templateDiskAttachmentsSet.List() {
		entry := item.(map[string]interface{})
		diskID := entry["disk_id"].(string)
		disk, err := ovirtclient.NewBuildableVMDiskParameters(ovirtclient.DiskID(diskID))
		if err != nil {
			diags = append(diags, errorToDiag("add disk to VM", err))
			return diags
		}
		if formatRaw, ok := entry["format"]; ok {
			disk, err = disk.WithFormat(ovirtclient.ImageFormat(formatRaw.(string)))
			if err != nil {
				diags = append(diags, errorToDiag("set format on disk", err))
				return diags
			}
		}
		if sparseRaw, ok := entry["sparse"]; ok {
			disk, err = disk.WithSparse(sparseRaw.(bool))
			if err != nil {
				diags = append(diags, errorToDiag("set sparse on disk", err))
				return diags
			}
		}
		disks[i] = disk
	}
	_, err := params.WithDisks(disks)
	if err != nil {
		diags = append(diags, errorToDiag("set disks on VM", err))
	}
	return diags
}

func handleVMMemoryPolicy(
	data *schema.ResourceData,
	params ovirtclient.BuildableVMParameters,
	diags diag.Diagnostics,
) diag.Diagnostics {
	addMemoryPolicy := false
	memoryPolicy := ovirtclient.NewMemoryPolicyParameters()
	maxMemory, ok := data.GetOk("maximum_memory")
	if ok {
		var err error
		_, err = memoryPolicy.WithMax(int64(maxMemory.(int)))
		if err != nil {
			diags = append(diags, errorToDiag("add maximum memory", err))
		} else {
			addMemoryPolicy = true
		}
	}
	ballooning, ok := data.GetOk("memory_ballooning")
	if ok {
		var err error
		_, err = memoryPolicy.WithBallooning(ballooning.(bool))
		if err != nil {
			diags = append(diags, errorToDiag("add ballooning", err))
		} else {
			addMemoryPolicy = true
		}
	}
	if addMemoryPolicy {
		params.WithMemoryPolicy(memoryPolicy)
	}
	return diags
}

func handleVMMemory(
	data *schema.ResourceData,
	params ovirtclient.BuildableVMParameters,
	diags diag.Diagnostics,
) diag.Diagnostics {
	memory, ok := data.GetOk("memory")
	if !ok {
		return diags
	}
	var err error
	_, err = params.WithMemory(int64(memory.(int)))
	if err != nil {
		diags = append(diags, errorToDiag("set memory", err))
	}
	return diags
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
	cpuMode, cpuModeOK := data.GetOk("cpu_mode")
	cpuCores, cpuCoresOK := data.GetOk("cpu_cores")
	cpuThreads, cpuThreadsOK := data.GetOk("cpu_threads")
	cpuSockets, cpuSocketsOK := data.GetOk("cpu_sockets")
	cpu := ovirtclient.NewVMCPUParams()
	cpuTopo := ovirtclient.NewVMCPUTopoParams()
	if cpuCoresOK {
		_, err := cpuTopo.WithCores(uint(cpuCores.(int)))
		if err != nil {
			diags = append(diags, errorToDiag("add CPU cores", err))
		}
	}
	if cpuThreadsOK {
		_, err := cpuTopo.WithThreads(uint(cpuThreads.(int)))
		if err != nil {
			diags = append(diags, errorToDiag("add CPU threads", err))
		}
	}
	if cpuSocketsOK {
		_, err := cpuTopo.WithSockets(uint(cpuSockets.(int)))
		if err != nil {
			diags = append(diags, errorToDiag("add CPU sockets", err))
		}
	}
	if cpuCoresOK || cpuThreadsOK || cpuSocketsOK {
		_, err := cpu.WithTopo(cpuTopo)
		if err != nil {
			diags = append(diags, errorToDiag("add CPU topology", err))
		}
	}
	if cpuModeOK {
		_, err := cpu.WithMode(ovirtclient.CPUMode(cpuMode.(string)))
		if err != nil {
			diags = append(diags, errorToDiag("add CPU mode", err))
		}
	}
	_, err := params.WithCPU(cpu)
	if err != nil {
		diags = append(diags, errorToDiag("add CPU", err))
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

func handleVMInitialization(
	data *schema.ResourceData,
	params ovirtclient.BuildableVMParameters,
	diags diag.Diagnostics,
) diag.Diagnostics {
	vmInitScript := ""
	vmHostname := ""
	useInit := false

	if hName, ok := data.GetOk("initialization_hostname"); ok {
		vmHostname = hName.(string)
		useInit = true
	}
	if hInitScript, ok := data.GetOk("initialization_custom_script"); ok {
		vmInitScript = hInitScript.(string)
		useInit = true
	}

	if useInit {
		_, err := params.WithInitialization(ovirtclient.NewInitialization(vmInitScript, vmHostname))
		if err != nil {
			diags = append(diags, errorToDiag("add Initialization parameters", err))
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

func cpuModeValues() []string {
	values := ovirtclient.CPUModeValues()
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = string(v)
	}
	return result
}
