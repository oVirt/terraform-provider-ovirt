// Copyright (C) 2017 Battelle Memorial Institute
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4"
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
				Default:  1,
			},
			"sockets": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"threads": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
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
							Default:  true,
						},
					},
				},
			},
		},
	}
}

func resourceVMCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)
	vmsService := conn.SystemService().VmsService()

	cluster, err := ovirtsdk4.NewClusterBuilder().
		Name(d.Get("cluster").(string)).Build()
	if err != nil {
		return err
	}

	template, err := ovirtsdk4.NewTemplateBuilder().
		Name(d.Get("template").(string)).Build()
	if err != nil {
		return err
	}

	cpuTopo := ovirtsdk4.NewCpuTopologyBuilder().
		Cores(int64(d.Get("cores").(int))).
		Threads(int64(d.Get("threads").(int))).
		Sockets(int64(d.Get("sockets").(int))).
		MustBuild()

	cpu, err := ovirtsdk4.NewCpuBuilder().
		Topology(cpuTopo).
		Build()
	if err != nil {
		return err
	}

	initialBuilder := ovirtsdk4.NewInitializationBuilder().
		AuthorizedSshKeys(d.Get("authorized_ssh_key").(string))

	numNetworks := d.Get("network_interface.#").(int)
	for i := 0; i < numNetworks; i++ {
		prefix := fmt.Sprintf("network_interface.%d", i)

		ncBuilder := ovirtsdk4.NewNicConfigurationBuilder().
			Name(d.Get(prefix + ".label").(string)).
			IpBuilder(
				ovirtsdk4.NewIpBuilder().
					Address(d.Get(prefix + ".ip_address").(string)).
					Netmask(d.Get(prefix + ".subnet_mask").(string)).
					Gateway(d.Get(prefix + ".gateway").(string))).
			BootProtocol(ovirtsdk4.BootProtocol(d.Get(prefix + ".boot_proto").(string))).
			OnBoot(d.Get(prefix + ".on_boot").(bool))
		initialBuilder.NicConfigurationsBuilderOfAny(*ncBuilder)
	}

	initialize, err := initialBuilder.Build()
	if err != nil {
		return err
	}

	resp, err := vmsService.Add().
		Vm(
			ovirtsdk4.NewVmBuilder().
				Name(d.Get("name").(string)).
				Cluster(cluster).
				Template(template).
				Cpu(cpu).
				Initialization(initialize).
				MustBuild()).
		Send()

	if err != nil {
		return err
	}
	newVM, ok := resp.Vm()
	if ok {
		d.SetId(newVM.MustId())
	}

	return resourceVMRead(d, meta)
}

func resourceVMUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceVMRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	getVmresp, err := conn.SystemService().VmsService().
		VmService(d.Id()).Get().Send()
	if err != nil {
		return err
	}

	vm, ok := getVmresp.Vm()

	if !ok {
		d.SetId("")
		return nil
	}
	d.Set("name", vm.MustName())
	d.Set("cores", vm.MustCpu().MustTopology().MustCores())
	d.Set("sockets", vm.MustCpu().MustTopology().MustSockets())
	d.Set("threads", vm.MustCpu().MustTopology().MustThreads())
	d.Set("authorized_ssh_key", vm.MustInitialization().MustAuthorizedSshKeys())

	// Use `conn.FollowLink` function to fetch cluster and template instance from a vm.
	// See: https://github.com/imjoey/go-ovirt/blob/master/examples/follow_vm_links.go.
	cluster, _ := conn.FollowLink(vm.MustCluster())
	if cluster, ok := cluster.(*ovirtsdk4.Cluster); ok {
		d.Set("cluster", cluster.MustName())
	}
	template, _ := conn.FollowLink(vm.MustTemplate())
	if template, ok := template.(*ovirtsdk4.Template); ok {
		d.Set("template", template.MustName())
	}

	return nil
}

func resourceVMDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	vmService := conn.SystemService().VmsService().VmService(d.Id())

	return resource.Retry(3*time.Minute, func() *resource.RetryError {
		getVMResp, err := vmService.Get().Send()
		if err != nil {
			return resource.RetryableError(err)
		}

		vm, ok := getVMResp.Vm()
		if !ok {
			d.SetId("")
			return nil
		}

		if vm.MustStatus() != ovirtsdk4.VMSTATUS_DOWN {
			_, err := vmService.Shutdown().Send()
			if err != nil {
				return resource.RetryableError(fmt.Errorf("Stop instance timeout and got an error: %v", err))
			}
		}
		//
		_, err = vmService.Remove().
			DetachOnly(true). // DetachOnly indicates without removing disks attachments
			Send()
		if err != nil {
			return resource.RetryableError(fmt.Errorf("Delete instalce timeout and got an error: %v", err))
		}

		return nil

	})
}
