// Package ovirt provides datasource for managing oVirt virtual machines
package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func (p *provider) vmDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: p.vmDataSourceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "oVirt ID of the VM. Either name or id must be specified.",
				ValidateDiagFunc: validateUUID,
			},
			"name": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "Human-readable name of the VM. Either name or id must be specified.",
				ValidateDiagFunc: validateNonEmpty,
			},
			"cluster_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cluster the VM belongs to.",
			},
			"template_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the base template for this VM.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the VM (e.g. up, down, etc.).",
			},
			"comment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Comment attached to the VM.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the VM.",
			},
			"memory": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Memory of the VM in bytes.",
			},
			"boot_devices": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Ordered list of boot devices for the VM.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		Description: "The ovirt_vm data source returns the ID and details of a virtual machine by its name or ID.",
	}
}

func (p *provider) vmDataSourceRead(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	client := p.client.WithContext(ctx)

	var id string
	if idVal, ok := data.GetOk("id"); ok {
		id = idVal.(string)
	}
	var name string
	if nameVal, ok := data.GetOk("name"); ok {
		name = nameVal.(string)
	}

	if id == "" && name == "" {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Missing VM identifier",
				Detail:   "Either 'name' or 'id' must be specified to look up a virtual machine.",
			},
		}
	}

	var vm ovirtclient.VM
	if id != "" {
		var err error
		vm, err = client.GetVM(ovirtclient.VMID(id))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Failed to get VM with ID %q", id),
					Detail:   err.Error(),
				},
			}
		}
		if name != "" && vm.Name() != name {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("VM name mismatch for ID %q", id),
					Detail:   fmt.Sprintf("VM with ID %q has name %q, but %q was specified.", id, vm.Name(), name),
				},
			}
		}
	} else {
		var err error
		vm, err = client.GetVMByName(name)
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Failed to get VM with name %q", name),
					Detail:   err.Error(),
				},
			}
		}
	}

	data.SetId(string(vm.ID()))

	diags := diag.Diagnostics{}
	diags = setResourceField(data, "name", vm.Name(), diags)
	diags = setResourceField(data, "cluster_id", string(vm.ClusterID()), diags)
	diags = setResourceField(data, "template_id", string(vm.TemplateID()), diags)
	diags = setResourceField(data, "status", string(vm.Status()), diags)
	diags = setResourceField(data, "comment", vm.Comment(), diags)
	diags = setResourceField(data, "description", vm.Description(), diags)
	diags = setResourceField(data, "memory", vm.Memory(), diags)
	if len(vm.OS().BootDevices()) > 0 {
		devices := make([]string, len(vm.OS().BootDevices()))
		for i, dev := range vm.OS().BootDevices() {
			devices[i] = string(dev)
		}
		diags = setResourceField(data, "boot_devices", devices, diags)
	}

	return diags
}
