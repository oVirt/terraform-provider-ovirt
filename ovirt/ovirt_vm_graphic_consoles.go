package ovirt

import (
    "context"

    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var vmGraphicConsoleSchema = map[string]*schema.Schema{
    "id": {
        Type:        schema.TypeString,
        Computed:    true,
        Description: "UUID of the graphics console.",
    },
}

var vmGraphicConsolesSchema = map[string]*schema.Schema{
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
    "console": {
        Type:        schema.TypeSet,
        Required:    true,
        Description: "The list of consoles that should be on this VM. If a console is not in this list it will be removed from the VM.",
        Elem: &schema.Resource{
            Schema: vmGraphicConsoleSchema,
        },
    },
}

func (p *provider) vmRemoveGraphicConsolesResource() *schema.Resource {
    return &schema.Resource{
        CreateContext: p.vmGraphicConsolesCreate,
        ReadContext:   p.vmGraphicConsolesRead,
        DeleteContext: p.vmGraphicConsolesDelete,
        Schema:        vmGraphicConsolesSchema,
        Description:   "The ovirt_vm_graphic_consoles controls all the graphic consoles of a VM.",
    }
}

func (p *provider) vmGraphicConsolesCreate(
    ctx context.Context,
    data *schema.ResourceData,
    i interface{},
) diag.Diagnostics {
    consoles := data.Get("console").(*schema.Set)

    if consoles.Len() != 0 {
        return diag.Diagnostics{
            diag.Diagnostic{
                Severity: diag.Error,
                Summary:  "Creating consoles is not supported",
                Detail:   "Currently, only removing all graphics consoles is supported.",
            },
        }
    }

}

func (p *provider) vmGraphicConsolesRead(
    ctx context.Context,
    data *schema.ResourceData,
    i interface{},
) diag.Diagnostics {
    // There is no way to re-read this.
    return nil
}

func (p *provider) vmGraphicConsolesDelete(
    ctx context.Context,
    data *schema.ResourceData,
    i interface{},
) diag.Diagnostics {
    vmID := data.Get("vm_id").(string)
    err := p.client.WithContext(ctx).AutoOptimizeVMCPUPinningSettings(vmID, false)
    return errorToDiags("auto-optimizing CPU pinning settings", err)
}
