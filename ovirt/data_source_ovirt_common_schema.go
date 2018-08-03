// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

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
