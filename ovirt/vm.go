package ovirt

import (
	"time"

	"github.com/EMSL-MSC/ovirtapi"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVM() *schema.Resource {
	return &schema.Resource{
		Create: resourceVMCreate,
		Read:   resourceVMRead,
		Update: resourceVMUpdate,
		Delete: resourceVMDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster": {
				Type:     schema.TypeString,
				Required: true,
			},
			"template": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceVMCreate(d *schema.ResourceData, meta interface{}) error {
	con := meta.(*ovirtapi.Connection)
	newVM := con.NewVM()
	newVM.Name = d.Get("name").(string)

	cluster := con.NewCluster()
	cluster.Name = d.Get("cluster").(string)
	newVM.Cluster = cluster

	template := con.NewTemplate()
	template.Name = d.Get("template").(string)
	newVM.Template = template

	err := newVM.Save()
	if err != nil {
		return err
	}

	for newVM.Status != "down" {
		time.Sleep(time.Second)
		newVM.Update()
	}

	newVM.Start("", "", "", "", "", nil)
	d.SetId(newVM.ID)
	return nil
}

func resourceVMUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}
func resourceVMRead(d *schema.ResourceData, meta interface{}) error {
	con := meta.(*ovirtapi.Connection)
	vm, err := con.GetVM(d.Id())

	if err != nil {
		d.SetId("")
		return nil
	}
	d.Set("name", vm.Name)
	d.Set("cluster", vm.Cluster.Name)
	d.Set("template", vm.Template.Name)
	return nil
}

func resourceVMDelete(d *schema.ResourceData, meta interface{}) error {
	con := meta.(*ovirtapi.Connection)
	vm, err := con.GetVM(d.Id())
	if err != nil {
		return nil
	}
	if vm.Status != "down" {
		vm.Stop("false")
	}
	for vm.Status != "down" {
		time.Sleep(time.Second)
		vm.Update()
	}
	return vm.Delete()
}
