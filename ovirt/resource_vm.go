package ovirt

import (
	"fmt"
	"strconv"
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
				Optional: true,
				Default:  "Default",
			},
			"template": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Blank",
			},
			"memory": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"cores": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"sockets": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"threads": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"authorized_ssh_key": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"network_interface": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"label": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"boot_proto": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"ip_address": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"subnet_mask": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"gateway": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"on_boot": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
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
        newVM.CPU = &ovirtapi.CPU{
            Topology: &ovirtapi.CPUTopology{
	        Cores: d.Get("cores").(int),
	        Sockets: d.Get("sockets").(int),
	        Threads: d.Get("threads").(int),
            },
        }
	newVM.Initialization = &ovirtapi.Initialization{}

	newVM.Initialization.AuthorizedSSHKeys = d.Get("authorized_ssh_key").(string)

	numNetworks := d.Get("network_interface.#").(int)
	NICConfigurations := make([]ovirtapi.NICConfiguration, numNetworks)
	for i := 0; i < numNetworks; i++ {
		prefix := fmt.Sprintf("network_interface.%d", i)
		_ = prefix
		NICConfigurations[i] = ovirtapi.NICConfiguration{
			IP: &ovirtapi.IP{
				Address: d.Get(prefix + ".ip_address").(string),
				Netmask: d.Get(prefix + ".subnet_mask").(string),
				Gateway: d.Get(prefix + ".gateway").(string),
			},
			BootProtocol: d.Get(prefix + ".boot_proto").(string),
			OnBoot:       strconv.FormatBool(d.Get(prefix + ".on_boot").(bool)),
			Name:         d.Get(prefix + ".label").(string),
		}
		if i == 0 {
			d.SetConnInfo(map[string]string{
				"host": d.Get(prefix + ".ip_address").(string),
			})
		}
	}
	newVM.Initialization.NICConfigurations = &ovirtapi.NICConfigurations{NICConfiguration: NICConfigurations}

	err := newVM.Save()
	if err != nil {
		return err
	}
	d.SetId(newVM.ID)

	for newVM.Status != "down" {
		time.Sleep(time.Second)
		newVM.Update()
	}

	err = newVM.Start("", "", "", "true", "", nil)
	if err != nil {
		newVM.Delete()
		return err
	}
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

	cluster, err := con.GetCluster(vm.Cluster.ID)
	if err != nil {
		d.SetId("")
		return nil
	}
	d.Set("cluster", cluster.Name)

	template, err := con.GetTemplate(vm.Template.ID)
	if err != nil {
		d.SetId("")
		return nil
	}
	d.Set("template", template.Name)
	d.Set("cores", vm.CPU.Topology.Cores)
	d.Set("sockets", vm.CPU.Topology.Sockets)
	d.Set("threads", vm.CPU.Topology.Threads)
	d.Set("authorized_ssh_key", vm.Initialization.AuthorizedSSHKeys)
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
