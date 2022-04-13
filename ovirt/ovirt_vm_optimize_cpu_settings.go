package ovirt

import (
    "context"

    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var vmOptimizeCPUSettingsSchema = map[string]*schema.Schema{
    "id": {
        Type:        schema.TypeString,
        Computed:    true,
        Description: "oVirt ID of the VM to be started.",
    },
    "vm_id": {
        Type:             schema.TypeString,
        Required:         true,
        Description:      "oVirt ID of the VM to be started.",
        ForceNew:         true,
        ValidateDiagFunc: validateUUID,
    },
}

func (p *provider) vmOptimizeCPUSettingsResource() *schema.Resource {
    return &schema.Resource{
        CreateContext: p.vmOptimizeCPUSettingsCreate,
        ReadContext:   p.vmOptimizeCPUSettingsRead,
        DeleteContext: p.vmOptimizeCPUSettingsDelete,
        Schema:        vmOptimizeCPUSettingsSchema,
        Description:   "The ovirt_vm_optimize_cpu_settings sets the CPU settings to automatically optimized for the specified VM.",
    }
}

func (p *provider) vmOptimizeCPUSettingsCreate(
    ctx context.Context,
    data *schema.ResourceData,
    i interface{},
) diag.Diagnostics {
    vmID := data.Get("vm_id").(string)
    err := p.client.WithContext(ctx).AutoOptimizeVMCPUPinningSettings(vmID, true)
    return errorToDiags("auto-optimizing CPU pinning settings", err)
}

func (p *provider) vmOptimizeCPUSettingsRead(
    ctx context.Context,
    data *schema.ResourceData,
    i interface{},
) diag.Diagnostics {
    // There is no way to re-read this.
    return nil
}

func (p *provider) vmOptimizeCPUSettingsDelete(
    ctx context.Context,
    data *schema.ResourceData,
    i interface{},
) diag.Diagnostics {
    vmID := data.Get("vm_id").(string)
    err := p.client.WithContext(ctx).AutoOptimizeVMCPUPinningSettings(vmID, false)
    return errorToDiags("auto-optimizing CPU pinning settings", err)
}
