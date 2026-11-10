// Package ovirt provides datasource for managing oVirt tags
package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func (p *provider) tagDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: p.tagDataSourceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "oVirt ID of the tag. Either id or name must be specified.",
				ValidateDiagFunc: validateUUID,
			},
			"name": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "Name of the tag. Either id or name must be specified.",
				ValidateDiagFunc: validateNonEmpty,
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the tag.",
			},
		},
		Description: "The ovirt_tag data source returns the ID and details of a tag by its name or ID.",
	}
}

func (p *provider) tagDataSourceRead(
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
				Summary:  "Missing tag identifier",
				Detail:   "Either 'name' or 'id' must be specified to look up a tag.",
			},
		}
	}

	var tag ovirtclient.Tag
	if id != "" {
		var err error
		tag, err = client.GetTag(ovirtclient.TagID(id))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Failed to get tag with ID %q", id),
					Detail:   err.Error(),
				},
			}
		}
		if name != "" && tag.Name() != name {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Tag name mismatch for ID %q", id),
					Detail:   fmt.Sprintf("Tag with ID %q has name %q, but %q was specified.", id, tag.Name(), name),
				},
			}
		}
	} else {
		tags, err := client.ListTags()
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to list tags",
					Detail:   err.Error(),
				},
			}
		}
		for _, candidate := range tags {
			if candidate.Name() == name {
				tag = candidate
				break
			}
		}
		if tag == nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Tag with name %q not found", name),
					Detail:   fmt.Sprintf("No tag with name %q was found.", name),
				},
			}
		}
	}

	data.SetId(string(tag.ID()))

	diags := diag.Diagnostics{}
	diags = setResourceField(data, "name", tag.Name(), diags)
	if desc := tag.Description(); desc != nil {
		diags = setResourceField(data, "description", *desc, diags)
	}

	return diags
}
