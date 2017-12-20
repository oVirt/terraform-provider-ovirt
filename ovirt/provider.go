package ovirt

import (
	"github.com/EMSL-MSC/ovirtapi"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns oVirt provider configuration
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Login username",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Login password",
				Sensitive:   true,
			},
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Ovirt server url",
			},
		},
		ConfigureFunc: ConfigureProvider,
		ResourcesMap: map[string]*schema.Resource{
			"ovirt_vm": resourceVM(),
		},
	}
}

func ConfigureProvider(d *schema.ResourceData) (interface{}, error) {
	return ovirtapi.NewConnection(d.Get("url").(string), d.Get("username").(string), d.Get("password").(string), false)
}
