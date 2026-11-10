// Package ovirt provides datasource for managing oVirt templates
package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func (p *provider) templateDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: p.templateDataSourceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "oVirt ID of the template. Either name or id must be specified.",
				ValidateDiagFunc: validateUUID,
			},
			"name": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "Name of the template. Either name or id must be specified.",
				ValidateDiagFunc: validateNonEmpty,
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the template.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the template (e.g. ok, locked).",
			},
		},
		Description: "The ovirt_template data source returns the ID and details of a template by its name or ID.",
	}
}

func (p *provider) templateDataSourceRead(
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
				Summary:  "Missing template identifier",
				Detail:   "Either 'name' or 'id' must be specified to look up a template.",
			},
		}
	}

	var template ovirtclient.Template
	if id != "" {
		var err error
		template, err = client.GetTemplate(ovirtclient.TemplateID(id))
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Failed to get template with ID %q", id),
					Detail:   err.Error(),
				},
			}
		}
		if name != "" && template.Name() != name {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Template name mismatch for ID %q", id),
					Detail:   fmt.Sprintf("Template with ID %q has name %q, but %q was specified.", id, template.Name(), name),
				},
			}
		}
	} else {
		var err error
		template, err = client.GetTemplateByName(name)
		if err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Failed to get template with name %q", name),
					Detail:   err.Error(),
				},
			}
		}
	}

	data.SetId(string(template.ID()))

	diags := diag.Diagnostics{}
	diags = setResourceField(data, "name", template.Name(), diags)
	diags = setResourceField(data, "description", template.Description(), diags)
	diags = setResourceField(data, "status", string(template.Status()), diags)

	return diags
}
