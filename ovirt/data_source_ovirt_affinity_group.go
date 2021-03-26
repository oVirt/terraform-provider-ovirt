// Copyright (C) 2021 Shantur Rathore <i@shantur.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func dataSourceOvirtAffinityGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOvirtAffinityGroupRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enforcing": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"positive": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"priority": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"hosts_rule": schemaTypeAffinityRule(),
			"vms_rule":   schemaTypeAffinityRule(),
		},
	}
}

func schemaTypeAffinityRule() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"enforcing": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"positive": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}

func dataSourceOvirtAffinityGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	clusterId := d.Get("cluster_id").(string)
	affinityGroupName := d.Get("name").(string)

	affinityGroupsResp, err := conn.SystemService().
		ClustersService().
		ClusterService(clusterId).
		AffinityGroupsService().
		List().
		Follow("cluster").
		Send()

	if err != nil {
		return err
	}

	affinityGroups, affinityGroupsOk := affinityGroupsResp.Groups()

	if affinityGroupsOk {

		for _, ag := range affinityGroups.Slice() {
			log.Printf("[DEBUG] affinityGroups in cluster (%s) name : %s and looking for %s", clusterId, ag.MustName(), affinityGroupName)
			if ag.MustName() == affinityGroupName {
				log.Printf("[DEBUG] affinityGroups in cluster (%s) found name : %s", clusterId, ag.MustName())
				return affinityGroupsDescriptionAtrributes(d, ag, meta)
			}
		}

		return fmt.Errorf("No affinityGroup with name %s found in cluster id %s", affinityGroupName, clusterId)

	} else {
		return fmt.Errorf("Error receiving affinity groups in cluster %s", clusterId)
	}

}

func affinityGroupsDescriptionAtrributes(d *schema.ResourceData, affinityGroup *ovirtsdk4.AffinityGroup, meta interface{}) error {
	desc, ok := affinityGroup.Description()
	if !ok {
		desc = ""
	}

	comment, ok := affinityGroup.Comment()
	if !ok {
		comment = ""
	}

	d.SetId(affinityGroup.MustId())

	err := d.Set("name", affinityGroup.MustName())
	if err != nil {
		return err
	}

	err = d.Set("cluster_id", affinityGroup.MustCluster().MustId())
	if err != nil {
		return err
	}

	err = d.Set("description", desc)
	if err != nil {
		return err
	}

	err = d.Set("comment", comment)
	if err != nil {
		return err
	}

	err = d.Set("enforcing", affinityGroup.MustEnforcing())
	if err != nil {
		return err
	}

	err = d.Set("hosts_rule", getAffinityRuleMap(affinityGroup.MustHostsRule()))
	if err != nil {
		return err
	}

	err = d.Set("positive", affinityGroup.MustPositive())
	if err != nil {
		return err
	}

	err = d.Set("priority", affinityGroup.MustPriority())
	if err != nil {
		return err
	}

	err = d.Set("vms_rule", getAffinityRuleMap(affinityGroup.MustVmsRule()))
	if err != nil {
		return err
	}

	// d.SetId(resource.UniqueId())
	return nil
}

func getAffinityRuleMap(affinityRule *ovirtsdk4.AffinityRule) map[string]string {
	log.Printf("[DEBUG] affinityRule : %s", strconv.FormatBool(affinityRule.MustEnabled()))
	return map[string]string{
		"enabled":   strconv.FormatBool(affinityRule.MustEnabled()),
		"enforcing": strconv.FormatBool(affinityRule.MustEnforcing()),
		"positive":  strconv.FormatBool(affinityRule.MustPositive()),
	}
}
