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
    "serial_console": {
        Type:        schema.TypeBool,
        Optional:    true,
        ForceNew:    true,
        Description: "Enable the serial console on the VM.",
    },
    "cpu_mode": {
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
        Description: "Set the CPU mode to 'host_passthrough', 'host_model', or 'custom'.",
        ValidateDiagFunc: validateEnum(
            []string{
                string(ovirtclient.CPUModeHostPassthrough),
                string(ovirtclient.CPUModeHostModel),
                string(ovirtclient.CPUModeCustom),
            },
        ),
    },
    "cpu_cores": {
        Type:             schema.TypeInt,
        Optional:         true,
        Default:          1,
        ForceNew:         true,
        RequiredWith:     []string{"cpu_sockets", "cpu_threads"},
        Description:      "Number of CPU cores to allocate to the VM. If set, cpu_threads and cpu_sockets must also be specified.",
        ValidateDiagFunc: validatePositiveInt,
    },
    "cpu_threads": {
        Type:             schema.TypeInt,
        Optional:         true,
        Default:          1,
        ForceNew:         true,
        RequiredWith:     []string{"cpu_sockets", "cpu_cores"},
        Description:      "Number of CPU threads to allocate to the VM. If set, cpu_cores and cpu_sockets must also be specified.",
        ValidateDiagFunc: validatePositiveInt,
    },
    "cpu_sockets": {
        Type:             schema.TypeInt,
        Optional:         true,
        Default:          1,
        ForceNew:         true,
        RequiredWith:     []string{"cpu_threads", "cpu_cores"},
        Description:      "Number of CPU sockets to allocate to the VM. If set, cpu_cores and cpu_threads must also be specified.",
        ValidateDiagFunc: validatePositiveInt,
    },
    "memory": {
        Type:             schema.TypeInt,
        Optional:         true,
        Default:          1073741824,
        ForceNew:         true,
        Description:      "Number of CPU sockets to allocate to the VM. If set, cpu_cores and cpu_threads must also be specified.",
        ValidateDiagFunc: validatePositiveInt,
    },
    "maximum_memory": {
        Type:             schema.TypeInt,
        Optional:         true,
        ForceNew:         true,
        Description:      "Maximum memory to set in the memory policy. Must be larger than 'memory'.",
        ValidateDiagFunc: validatePositiveInt,
    },
    "memory_ballooning": {
        Type:        schema.TypeBool,
        Optional:    true,
        ForceNew:    true,
        Description: "Enable or disable memory ballooning.",
    },
    "os_type": {
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
        Description: "Sets the operating system type. See the QEMU documentation for supported strings.",
    },
    "initialization_custom_script": {
        Type:     schema.TypeString,
        Optional: true,
        ForceNew: true,
    },
    "initialization_hostname": {
        Type:     schema.TypeString,
        Optional: true,
        ForceNew: true,
    },
    "template_disk_attachment_override": {
        Type:     schema.TypeSet,
        Optional: true,
        ForceNew: true,
        Elem: &schema.Resource{
            Schema: map[string]*schema.Schema{
                "disk_id": {
                    Type:             schema.TypeString,
                    Required:         true,
                    ForceNew:         true,
                    ValidateDiagFunc: validateUUID,
                },
                "format": {
                    Type:             schema.TypeString,
                    Optional:         true,
                    ForceNew:         true,
                    ValidateDiagFunc: validateFormat,
                },
                "sparse": {
                    Type:     schema.TypeBool,
                    Optional: true,
                    ForceNew: true,
                },
            },
        },
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
    if mem, ok := data.GetOk("memory"); ok {
        _, err := params.WithMemory(int64(mem.(int)))
        if err != nil {
            errorToDiags("add memory", err)
        }
    }

    if osType, ok := data.GetOk("os_type"); ok {
        os, err := ovirtclient.NewVMOSParameters().WithType(osType.(string))
        if err != nil {
            return errorToDiags("adding OS type to VM", err)
        }
        params.WithOS(os)
    }

    customScript := ""
    hostName := ""
    if initializationContent, ok := data.GetOk("initialization_custom_script"); ok {
        customScript = initializationContent.(string)
    }
    if hostNameContent, ok := data.GetOk("initialization_hostname"); ok {
        hostName = hostNameContent.(string)
    }
    if customScript != "" || hostName != "" {
        params.MustWithInitializationParameters(customScript, hostName)
    }

    if tplDiskOverrideRaw, ok := data.GetOk("template_disk_attachment_override"); ok {
        tplDiskOverrideList := tplDiskOverrideRaw.(*schema.Set).List()
        diskParams := make([]ovirtclient.OptionalVMDiskParameters, len(tplDiskOverrideList))
        for i, raw := range tplDiskOverrideList {
            tplDiskOverride := raw.(map[string]interface{})
            override, err := ovirtclient.NewBuildableVMDiskParameters(tplDiskOverride["disk_id"].(string))
            if err != nil {
                return errorToDiags(fmt.Sprintf("process disk override %d", i), err)
            }
            if format, ok := tplDiskOverride["format"]; ok {
                if _, err := override.WithFormat(ovirtclient.ImageFormat(format.(string))); err != nil {
                    return errorToDiags(fmt.Sprintf("process disk override %d", i), err)
                }
            }
            if sparse, ok := tplDiskOverride["sparse"]; ok {
                if _, err := override.WithSparse(sparse.(bool)); err != nil {
                    return errorToDiags(fmt.Sprintf("process disk override %d", i), err)
                }
            }
            diskParams[i] = override
        }
        if _, err := params.WithDisks(diskParams); err != nil {
            return errorToDiags("process disk overrides", err)
        }
    }

    vm, err := p.client.WithContext(ctx).CreateVM(
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
    id := data.Id()
    vm, err := p.client.WithContext(ctx).GetVM(id)
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
    diags = setResourceField(data, "memory", vm.Memory(), diags)
    diags = setResourceField(data, "cpu_cores", vm.CPU().Topo().Cores(), diags)
    diags = setResourceField(data, "cpu_threads", vm.CPU().Topo().Threads(), diags)
    diags = setResourceField(data, "cpu_sockets", vm.CPU().Topo().Sockets(), diags)
    if _, ok := data.GetOk("os_type"); ok || vm.OS().Type() != "other" {
        diags = setResourceField(data, "os_type", vm.OS().Type(), diags)
    }
    if _, ok := data.GetOk("initialization_custom_script"); ok || vm.Initialization().CustomScript() != "" {
        diags = setResourceField(data, "initialization_custom_script", vm.Initialization().CustomScript(), diags)
    }
    if _, ok := data.GetOk("initialization_hostname"); ok || vm.Initialization().HostName() != "" {
        diags = setResourceField(data, "initialization_hostname", vm.Initialization().HostName(), diags)
    }
    if _, ok := data.GetOk("cpu_mode"); ok || vm.CPU().Mode() != nil {
        mode := ""
        if vm.CPU().Mode() != nil {
            mode = string(*vm.CPU().Mode())
        }
        diags = setResourceField(data, "cpu_mode", mode, diags)
    }
    return diags
}

func (p *provider) vmDelete(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
    if err := p.client.WithContext(ctx).RemoveVM(data.Id()); err != nil {
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

    vm, err := p.client.WithContext(ctx).UpdateVM(data.Id(), params)
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
    vm, err := p.client.WithContext(ctx).GetVM(data.Id())
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
