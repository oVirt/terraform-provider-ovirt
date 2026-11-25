// Package ovirt provides datasource for listing and searching oVirt virtual machines
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
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter virtual machines by exact name.",
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter virtual machines by status (e.g. up, down).",
			},
			"cluster_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Filter virtual machines by cluster ID.",
				ValidateDiagFunc: validateUUID,
			},
			"fail_on_empty": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Fail if no virtual machines matching the criteria were found.",
			},
			"vms": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of virtual machines matching the search criteria.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "oVirt ID of the VM.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the VM.",
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
							Description: "Status of the VM.",
						},
						"comment": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Comment attached to the VM.",
						},
						"memory": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Memory of the VM in bytes.",
						},
					},
				},
			},
		},
		Description: "The ovirt_vms data source returns a list of virtual machines matching the given search criteria.",
	}
}

func (p *provider) vmsDataSourceRead(
	ctx context.Context,
	data *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	client := p.client.WithContext(ctx)

	var filterName string
	if nameVal, ok := data.GetOk("name"); ok {
		filterName = nameVal.(string)
	}
	var filterStatus string
	if statusVal, ok := data.GetOk("status"); ok {
		filterStatus = statusVal.(string)
	}
	var filterClusterID string
	if clusterVal, ok := data.GetOk("cluster_id"); ok {
		filterClusterID = clusterVal.(string)
	}

	allVMs, err := client.ListVMs()
	if err != nil {
		return errorToDiags("list VMs", err)
	}

	result := make([]map[string]interface{}, 0)
	for _, vm := range allVMs {
		if filterName != "" && vm.Name() != filterName {
			continue
		}
		if filterStatus != "" && string(vm.Status()) != filterStatus {
			continue
		}
		if filterClusterID != "" && string(vm.ClusterID()) != filterClusterID {
			continue
		}

		result = append(result, map[string]interface{}{
			"id":          string(vm.ID()),
			"name":        vm.Name(),
			"cluster_id":  string(vm.ClusterID()),
			"template_id": string(vm.TemplateID()),
			"status":      string(vm.Status()),
			"comment":     vm.Comment(),
			"memory":      vm.Memory(),
		})
	}

	if data.Get("fail_on_empty").(bool) && len(result) == 0 {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "No virtual machines found",
				Detail:   "No virtual machines found matching the specified criteria.",
			},
		}
	}

	if err := data.Set("vms", result); err != nil {
		return errorToDiags("set vms", err)
	}

	data.SetId(fmt.Sprintf("vms-%s-%s-%s", filterName, filterStatus, filterClusterID))

	return nil
}
