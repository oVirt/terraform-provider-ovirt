package ovirt

import (
	"strconv"

	"github.com/EMSL-MSC/ovirtapi"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceDisk() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDiskRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"format": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"storage_domain_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"shareable": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"sparse": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func dataSourceDiskRead(d *schema.ResourceData, meta interface{}) error {
	con := meta.(*ovirtapi.Connection)
	disk, err := con.GetDisk(d.Id())
	if err != nil {
		d.SetId("")
		return nil
	}

	d.Set("name", disk.Name)
	d.Set("size", disk.ProvisionedSize)
	d.Set("format", disk.Format)
	d.Set("storage_domain_id", disk.StorageDomains.StorageDomain[0].ID)
	shareable, _ := strconv.ParseBool(disk.Shareable)
	d.Set("shareable", shareable)
	sparse, _ := strconv.ParseBool(disk.Sparse)
	d.Set("sparse", sparse)
	return nil
}
