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

func resourceOvirtDataCenter() *schema.Resource {
	return &schema.Resource{
		Create: resourceOvirtDataCenterCreate,
		Read:   resourceOvirtDataCenterRead,
		Update: resourceOvirtDataCenterUpdate,
		Delete: resourceOvirtDataCenterDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			// This field identifies whether the datacenter uses local storage
			"local": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: false,
			},
		},
	}
}

func resourceOvirtDataCenterCreate(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*ovirtsdk4.Connection)
	name := d.Get("name").(string)
	local := d.Get("local").(bool)

	//Name and Local are required when create a datacenter
	datacenterbuilder := ovirtsdk4.NewDataCenterBuilder().Name(name).Local(local)

	// Check if has description
	_, ok := d.GetOkExists("description")
	if ok {
		description := d.Get("description").(string)
		datacenterbuilder = datacenterbuilder.Description(description)
	}

	datacenter, err := datacenterbuilder.Build()
	if err != nil {
		return err
	}

	addResp, err := conn.SystemService().DataCentersService().Add().DataCenter(datacenter).Send()
	if err != nil {
		return err
	}

	d.SetId(addResp.MustDataCenter().MustId())
	return resourceOvirtDataCenterRead(d, meta)

}

func resourceOvirtDataCenterUpdate(d *schema.ResourceData, meta interface{}) error {

	var ok bool
	conn := meta.(*ovirtsdk4.Connection)
	datacenterService := conn.SystemService().DataCentersService().DataCenterService(d.Id())
	datacenterBuilder := ovirtsdk4.NewDataCenterBuilder()

    _, ok = d.GetOkExists("name")
    if ok {
		if d.HasChange("name") {
			name := d.Get("name").(string)
			datacenterBuilder.Name(name)
		}
	}else{
		return fmt.Errorf("DataCenter's name don't not exist!")
	}

	_, ok = d.GetOkExists("description")
	if ok && d.HasChange("description") {
		description := d.Get("description").(string)
		datacenterBuilder.Description(description)
	}

    _, ok = d.GetOkExists("local")
    if ok {
		if d.HasChange("local") {
			local := d.Get("local").(bool)
			datacenterBuilder.Local(local)
		}
	}else{
		return fmt.Errorf("DataCenter's local don't not exist!")
	}

	datacenter, err := datacenterBuilder.Build()
	if err != nil {
		return err
	}

	_, err = datacenterService.Update().DataCenter(datacenter).Send()

	if err != nil {
		return err
	}

	return resourceOvirtDataCenterRead(d, meta)
}

func resourceOvirtDataCenterRead(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*ovirtsdk4.Connection)
	getDataCenterResp, err := conn.SystemService().DataCentersService().
		DataCenterService(d.Id()).Get().Send()
	if err != nil {
		return err
	}

	datacenter, ok := getDataCenterResp.DataCenter()
	if !ok {
		d.SetId("")
		return nil
	}

	d.Set("name", datacenter.MustName())
	d.Set("local", datacenter.MustLocal())

	description, ok := datacenter.Description()
	if ok {
		d.Set("description", description)
	}

	return nil
}

func resourceOvirtDataCenterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	_, err := conn.SystemService().DataCentersService().
		DataCenterService(d.Id()).Remove().Send()
	if err != nil {
		return err
	}
	return nil
}
