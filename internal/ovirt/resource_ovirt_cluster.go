//nolint:revive
package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

var clusterSchema = map[string]*schema.Schema{
	"id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "oVirt ID of this cluster.",
	},
	"name": {
		Type:             schema.TypeString,
		Required:         true,
		Description:      "User-provided name for the cluster.",
		ValidateDiagFunc: validateNonEmpty,
	},
	"datacenter_id": {
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		Description:      "Datacenter to create this cluster in.",
		ValidateDiagFunc: validateUUID,
	},
	"cpu_type": {
		Type:             schema.TypeString,
		Required:         true,
		Description:      "CPU type/model for the cluster (e.g. 'Intel SandyBridge Family').",
		ValidateDiagFunc: validateNonEmpty,
	},
	"cpu_arch": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "CPU architecture for the cluster (e.g. 'x86_64').",
	},
	"description": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "User-provided description for the cluster.",
	},
	"comment": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "User-provided comment for the cluster.",
	},
	"ballooning": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "Enable memory ballooning for VMs in this cluster.",
	},
	"memory_policy_over_commit_percent": {
		Type:             schema.TypeInt,
		Optional:         true,
		Computed:         true,
		ValidateDiagFunc: validatePositiveInt,
		Description:      "Memory overcommit percentage for the cluster (e.g. 100, 150, 200).",
	},
}

func (p *provider) clusterResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: p.clusterCreate,
		ReadContext:   p.clusterRead,
		UpdateContext: p.clusterUpdate,
		DeleteContext: p.clusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema:      clusterSchema,
		Description: "The ovirt_cluster resource creates clusters in oVirt.",
	}
}

func (p *provider) clusterCreate(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	name := data.Get("name").(string)
	dcID := data.Get("datacenter_id").(string)
	cpuType := data.Get("cpu_type").(string)

	params := ovirtclient.NewCreateClusterParams()
	if desc, ok := data.GetOk("description"); ok {
		var err error
		params, err = params.WithDescription(desc.(string))
		if err != nil {
			return errorToDiags("set description", err)
		}
	}
	if comment, ok := data.GetOk("comment"); ok {
		var err error
		params, err = params.WithComment(comment.(string))
		if err != nil {
			return errorToDiags("set comment", err)
		}
	}
	if arch, ok := data.GetOk("cpu_arch"); ok {
		var err error
		params, err = params.WithCPUArchitecture(arch.(string))
		if err != nil {
			return errorToDiags("set cpu_arch", err)
		}
	}
	if b, ok := data.GetOkExists("ballooning"); ok {
		var err error
		params, err = params.WithBallooning(b.(bool))
		if err != nil {
			return errorToDiags("set ballooning", err)
		}
	}
	if overcommit, ok := data.GetOk("memory_policy_over_commit_percent"); ok {
		var err error
		params, err = params.WithMemoryOverCommitPercent(int64(overcommit.(int)))
		if err != nil {
			return errorToDiags("set memory_policy_over_commit_percent", err)
		}
	}

	c, err := client.CreateCluster(ovirtclient.DatacenterID(dcID), name, cpuType, params)
	if err != nil {
		return errorToDiags(fmt.Sprintf("create cluster %s", name), err)
	}

	return clusterToData(c, data)
}

func (p *provider) clusterRead(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	c, err := client.GetCluster(ovirtclient.ClusterID(data.Id()))
	if err != nil {
		if isNotFound(err) {
			data.SetId("")
			return nil
		}
		return errorToDiags(fmt.Sprintf("get cluster %s", data.Id()), err)
	}
	return clusterToData(c, data)
}

func (p *provider) clusterUpdate(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)

	params := ovirtclient.NewUpdateClusterParams()
	if data.HasChange("name") {
		var err error
		params, err = params.WithName(data.Get("name").(string))
		if err != nil {
			return errorToDiags("set name", err)
		}
	}
	if data.HasChange("cpu_type") {
		var err error
		params, err = params.WithCPUType(data.Get("cpu_type").(string))
		if err != nil {
			return errorToDiags("set cpu_type", err)
		}
	}
	if data.HasChange("cpu_arch") {
		var err error
		params, err = params.WithCPUArchitecture(data.Get("cpu_arch").(string))
		if err != nil {
			return errorToDiags("set cpu_arch", err)
		}
	}
	if data.HasChange("description") {
		var err error
		params, err = params.WithDescription(data.Get("description").(string))
		if err != nil {
			return errorToDiags("set description", err)
		}
	}
	if data.HasChange("comment") {
		var err error
		params, err = params.WithComment(data.Get("comment").(string))
		if err != nil {
			return errorToDiags("set comment", err)
		}
	}
	if data.HasChange("ballooning") {
		if b, ok := data.GetOkExists("ballooning"); ok {
			var err error
			params, err = params.WithBallooning(b.(bool))
			if err != nil {
				return errorToDiags("set ballooning", err)
			}
		}
	}
	if data.HasChange("memory_policy_over_commit_percent") {
		if overcommit, ok := data.GetOk("memory_policy_over_commit_percent"); ok {
			var err error
			params, err = params.WithMemoryOverCommitPercent(int64(overcommit.(int)))
			if err != nil {
				return errorToDiags("set memory_policy_over_commit_percent", err)
			}
		}
	}

	c, err := client.UpdateCluster(ovirtclient.ClusterID(data.Id()), params)
	if err != nil {
		return errorToDiags(fmt.Sprintf("update cluster %s", data.Id()), err)
	}
	return clusterToData(c, data)
}

func (p *provider) clusterDelete(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	if err := client.RemoveCluster(ovirtclient.ClusterID(data.Id())); err != nil {
		if isNotFound(err) {
			data.SetId("")
			return nil
		}
		return errorToDiags(fmt.Sprintf("remove cluster %s", data.Id()), err)
	}
	data.SetId("")
	return nil
}

func clusterToData(c ovirtclient.Cluster, data *schema.ResourceData) diag.Diagnostics {
	diags := diag.Diagnostics{}
	data.SetId(string(c.ID()))
	diags = appendDiags(diags, "set name on ovirt_cluster", data.Set("name", c.Name()))
	diags = appendDiags(diags, "set datacenter_id on ovirt_cluster", data.Set("datacenter_id", string(c.DatacenterID())))
	diags = appendDiags(diags, "set cpu_type on ovirt_cluster", data.Set("cpu_type", c.CPUType()))
	if arch := c.CPUArchitecture(); arch != nil {
		diags = appendDiags(diags, "set cpu_arch on ovirt_cluster", data.Set("cpu_arch", *arch))
	}
	if desc := c.Description(); desc != nil {
		diags = appendDiags(diags, "set description on ovirt_cluster", data.Set("description", *desc))
	} else {
		diags = appendDiags(diags, "clear description on ovirt_cluster", data.Set("description", nil))
	}
	if comment := c.Comment(); comment != nil {
		diags = appendDiags(diags, "set comment on ovirt_cluster", data.Set("comment", *comment))
	} else {
		diags = appendDiags(diags, "clear comment on ovirt_cluster", data.Set("comment", nil))
	}
	if b := c.Ballooning(); b != nil {
		diags = appendDiags(diags, "set ballooning on ovirt_cluster", data.Set("ballooning", *b))
	}
	if overcommit := c.MemoryOverCommitPercent(); overcommit != nil {
		diags = appendDiags(diags, "set memory_policy_over_commit_percent on ovirt_cluster", data.Set("memory_policy_over_commit_percent", int(*overcommit)))
	}
	return diags
}
