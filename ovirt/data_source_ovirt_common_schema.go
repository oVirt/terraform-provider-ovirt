package ovirt

import "github.com/hashicorp/terraform/helper/schema"

func dataSourceSearchSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"criteria": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"max": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"case_sensitive": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
			},
		},
	}
}
