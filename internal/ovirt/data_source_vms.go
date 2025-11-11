package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (p *provider) vmsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: p.vmsDataSourceRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Name of the VM to look for",
				ValidateDiagFunc: validateNonEmpty,
			},
			"fail_on_empty": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Fail if no VMs with the given name were found.",
			},
			"vms": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "oVirt identifier for the VM",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Current status of the VM (up, down, etc.).",
						},
						"ips": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of IP addresses reported for this VM (empty if none).",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
		Description: `Search oVirt VMs by name.`,
	}
}

func (p *provider) vmsDataSourceRead(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	vms, err := client.ListVMs()
	if err != nil {
		return errorToDiags("list VMs", err)
	}
	name := data.Get("name").(string)
	var result []map[string]interface{}
	for _, vm := range vms {
		if vm.Name() == name {
			ipMap, err := client.GetVMIPAddresses(vm.ID(), nil)
			ips := []string{}
			if err == nil && ipMap != nil {
				for _, ipList := range ipMap {
					for _, ip := range ipList {
						if ip != nil {
							ips = append(ips, ip.String())
						}
					}
				}
			}

			result = append(result, map[string]interface{}{
				"id":     vm.ID(),
				"status": vm.Status(),
				"ips":    ips,
			})
		}
	}
	data.SetId(name)
	if err := data.Set("vms", result); err != nil {
		return errorToDiags("set VMs", err)
	}
	if data.Get("fail_on_empty").(bool) && len(result) == 0 {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "No VM found",
				Detail:   fmt.Sprintf("No VM with the name %s found.", name),
			},
		}
	}
	return nil
}
