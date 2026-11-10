// Package ovirt provides datasource for managing oVirt cluster
package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func (p *provider) clusterDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: p.clusterDataSourceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "oVirt ID of the cluster. Either name or id must be specified.",
				ValidateDiagFunc: validateUUID,
			},
			"name": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "Human-readable name of the cluster. Either name or id must be specified.",
				ValidateDiagFunc: validateNonEmpty,
			},
		},
		Description: "The ovirt_cluster data source returns the ID and details of a cluster by its name or ID.",
	}
}

func (p *provider) clusterDataSourceRead(
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
				Summary:  "Missing cluster identifier",
				Detail:   "Either 'name' or 'id' must be specified to look up a cluster.",
			},
		}
	}

	var cluster ovirtclient.Cluster
	if id != "" {
		var err error
		cluster, err = client.GetCluster(ovirtclient.ClusterID(id))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Failed to get cluster with ID %q", id),
					Detail:   err.Error(),
				},
			}
		}
		if name != "" && cluster.Name() != name {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Cluster name mismatch for ID %q", id),
					Detail:   fmt.Sprintf("Cluster with ID %q has name %q, but %q was specified.", id, cluster.Name(), name),
				},
			}
		}
	} else {
		clusters, err := client.ListClusters()
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to list clusters",
					Detail:   err.Error(),
				},
			}
		}

		var matched []ovirtclient.Cluster
		for _, c := range clusters {
			if c.Name() == name {
				matched = append(matched, c)
			}
		}

		if len(matched) == 0 {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Cluster with name %q not found", name),
					Detail:   fmt.Sprintf("No cluster found with name %q in the oVirt engine.", name),
				},
			}
		}
		cluster = matched[0]
	}

	data.SetId(string(cluster.ID()))

	diags := diag.Diagnostics{}
	diags = setResourceField(data, "name", cluster.Name(), diags)

	return diags
}
