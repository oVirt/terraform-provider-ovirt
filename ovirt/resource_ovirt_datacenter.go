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
		Importer: &schema.ResourceImporter{
			State: resourceOvirtDataCenterImportState,
		},
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
	if description, ok := d.GetOkExists("description"); ok {
		datacenterbuilder = datacenterbuilder.Description(description.(string))
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

	conn := meta.(*ovirtsdk4.Connection)
	datacenterService := conn.SystemService().DataCentersService().DataCenterService(d.Id())
	datacenterBuilder := ovirtsdk4.NewDataCenterBuilder()

	if name, ok := d.GetOkExists("name"); ok {
		if d.HasChange("name") {
			datacenterBuilder.Name(name.(string))
		}
	} else {
		return fmt.Errorf("datacenter's name does not exist")
	}

	if description, ok := d.GetOkExists("description"); ok && d.HasChange("description") {
		datacenterBuilder.Description(description.(string))
	}

	if local, ok := d.GetOkExists("local"); ok {
		if d.HasChange("local") {
			datacenterBuilder.Local(local.(bool))
		}
	} else {
		return fmt.Errorf("datacenter's local does not exist")
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

	if description, ok := datacenter.Description(); ok {
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

func resourceOvirtDataCenterImportState(d *schema.ResourceData,
	meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*ovirtsdk4.Connection)

	//	if resp.DataCenter
	resp, err := conn.SystemService().DataCentersService().DataCenterService(d.Id()).Get().Send()
	if err != nil {
		return nil, err
	}
	datacenter, ok := resp.DataCenter()
	if !ok {
		d.SetId("")
		return nil, nil
	}
	d.Set("name", datacenter.MustName())
	d.Set("local", datacenter.MustLocal())
	if description, ok := datacenter.Description(); ok {
		d.Set("description", description)
	}
	return []*schema.ResourceData{d}, nil
}
