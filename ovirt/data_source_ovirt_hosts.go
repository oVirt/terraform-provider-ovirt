// Copyright (C) 2017 Battelle Memorial Institute
// Copyright (C) 2018 Chunguang Wu <chokko@126.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4"
)

func dataSourceOvirtHosts() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOvirtHostsRead,
		Schema: map[string]*schema.Schema{
			"search": dataSourceSearchSchema(),
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.ValidateRegexp,
			},

			"hosts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceOvirtHostsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	hostsReq := conn.SystemService().HostsService().List()

	search, searchOK := d.GetOk("search")
	nameRegex, nameRegexOK := d.GetOk("name_regex")

	if searchOK {
		searchMap := search.(map[string]interface{})
		searchCriteria, searchCriteriaOK := searchMap["criteria"]
		searchMax, searchMaxOK := searchMap["max"]
		searchCaseSensitive, searchCaseSensitiveOK := searchMap["case_sensitive"]
		if !searchCriteriaOK && !searchMaxOK && !searchCaseSensitiveOK {
			return fmt.Errorf("One of criteria or max or case_sensitive in search must be assigned")
		}

		if searchCriteriaOK {
			hostsReq.Search(searchCriteria.(string))
		}
		if searchMaxOK {
			maxInt, err := strconv.ParseInt(searchMax.(string), 10, 64)
			if err != nil || maxInt < 1 {
				return fmt.Errorf("search.max must be a positive int")
			}
			hostsReq.Max(maxInt)
		}
		if searchCaseSensitiveOK {
			csBool, err := strconv.ParseBool(searchCaseSensitive.(string))
			if err != nil {
				return fmt.Errorf("search.case_sensitive must be true or false")
			}
			hostsReq.CaseSensitive(csBool)
		}
	}
	hostResp, err := hostsReq.Send()
	if err != nil {
		return err
	}
	hosts, ok := hostResp.Hosts()
	if !ok || len(hosts.Slice()) == 0 {
		return fmt.Errorf("your query returned no results, please change your search criteria and try again")
	}

	var filterdHosts []*ovirtsdk4.Host
	if nameRegexOK {
		r := regexp.MustCompile(nameRegex.(string))
		for _, c := range hosts.Slice() {
			if r.MatchString(c.MustName()) {
				filterdHosts = append(filterdHosts, c)
			}
		}
	} else {
		filterdHosts = hosts.Slice()[:]
	}

	if len(filterdHosts) == 0 {
		return fmt.Errorf("your query returned no results, please change your search criteria and try again")
	}

	return hostsDescriptionAttributes(d, filterdHosts, meta)
}

func hostsDescriptionAttributes(d *schema.ResourceData, hosts []*ovirtsdk4.Host, meta interface{}) error {
	var s []map[string]interface{}
	for _, v := range hosts {

		mapping := map[string]interface{}{
			"id":   v.MustId(),
			"name": v.MustName(),
		}
		s = append(s, mapping)
	}

	d.SetId(resource.UniqueId())
	if err := d.Set("hosts", s); err != nil {
		return err
	}

	return nil
}
