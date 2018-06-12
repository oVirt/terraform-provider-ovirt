// Copyright (C) 2017 Battelle Memorial Institute
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4"
)

func resourceDataCenter() *schema.Resource {
	return &schema.Resource{
		Create: resourceDataCenterCreate,
		Read:   resourceDataCenterRead,
		Delete: resourceDataCenterDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"local": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDataCenterCreate(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*ovirtsdk4.Connection)

	DataCenterBuilder := ovirtsdk4.NewDataCenterBuilder().
		Name(d.Get("name").(string)).
		Description(d.Get("description").(string)).
		Local(d.Get("local").(bool)).
		MustBuild()

	DataCenter := DataCenterBuilder

	addResp, err := conn.SystemService().DataCentersService().Add().DataCenter(DataCenter).Send()
	if err != nil {
		return err
	}

	d.SetId(addResp.MustDataCenter().MustId())
	return nil

}

func resourceDataCenterRead(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*ovirtsdk4.Connection)
	getDataCeneterresp := conn.SystemService().DataCentersService()

	datacentersResponse, err := getDataCeneterresp.List().Send()

	if err != nil {
		return nil
	}

	if datacenters, ok := datacentersResponse.DataCenters(); ok {
		for _, dc := range datacenters.Slice() {
			fmt.Printf("DataCenter: ")
			if dcName, ok := dc.Name(); ok {
				fmt.Printf(" name: %v", dcName)
			}
			if dcID, ok := dc.Id(); ok {
				fmt.Printf(" id: %v", dcID)
			}
			fmt.Printf("  Supported versions are: ")
			if svs, ok := dc.SupportedVersions(); ok {
				for _, sv := range svs.Slice() {
					if major, ok := sv.Major(); ok {
						fmt.Printf(" Major: %v", major)
					}
					if minor, ok := sv.Minor(); ok {
						fmt.Printf(" Minor: %v", minor)
					}
				}
			}
			fmt.Println("")
		}
	}

	return nil

}

func resourceDataCenterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	_, err := conn.SystemService().DataCentersService().
		DataCenterService(d.Id()).Remove().Send()
	if err != nil {
		return err
	}
	return nil
}
