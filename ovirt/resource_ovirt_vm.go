// Copyright (C) 2017 Battelle Memorial Institute
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4"
)

func resourceOvirtVM() *schema.Resource {
	return &schema.Resource{
		Create: resourceOvirtVMCreate,
		Read:   resourceOvirtVMRead,
		Update: resourceOvirtVMUpdate,
		Delete: resourceOvirtVMDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"template": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Blank",
				ForceNew: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"cores": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
				ForceNew: true,
			},
			"sockets": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
				ForceNew: true,
			},
			"threads": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
				ForceNew: true,
			},
			"attached_disk": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk_attachment_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"disk_id": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"active": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"bootable": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"interface": &schema.Schema{
							Type: schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{
								string(ovirtsdk4.DISKINTERFACE_IDE),
								string(ovirtsdk4.DISKINTERFACE_SPAPR_VSCSI),
								string(ovirtsdk4.DISKINTERFACE_VIRTIO),
								string(ovirtsdk4.DISKINTERFACE_VIRTIO_SCSI),
							}, false),
							Required: true,
						},
						"logical_name": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"pass_discard": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
						},
						"read_only": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"use_scsi_reservation": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"initialization": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"timezone": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"user_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"custom_script": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dns_servers": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dns_search": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"nic_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"label": &schema.Schema{
										Type:     schema.TypeString,
										Required: true,
									},
									"boot_proto": &schema.Schema{
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(ovirtsdk4.BOOTPROTOCOL_AUTOCONF),
											string(ovirtsdk4.BOOTPROTOCOL_DHCP),
											string(ovirtsdk4.BOOTPROTOCOL_NONE),
											string(ovirtsdk4.BOOTPROTOCOL_STATIC),
										}, false),
									},
									"address": &schema.Schema{
										Type:     schema.TypeString,
										Optional: true,
									},
									"netmask": &schema.Schema{
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
						"authorized_ssh_key": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
					},
				},
			},
			"vnic": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vnic_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"vnic_profile_id": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceOvirtVMCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)
	vmsService := conn.SystemService().VmsService()

	vmBuilder := ovirtsdk4.NewVmBuilder().
		Name(d.Get("name").(string))

	cluster, err := ovirtsdk4.NewClusterBuilder().
		Id(d.Get("cluster_id").(string)).Build()
	if err != nil {
		return err
	}
	vmBuilder.Cluster(cluster)

	template, err := ovirtsdk4.NewTemplateBuilder().
		Name(d.Get("template").(string)).Build()
	if err != nil {
		return err
	}
	vmBuilder.Template(template)

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
	vmBuilder.Cpu(cpu)

	if v, ok := d.GetOk("initialization"); ok {
		initialization, err := expandOvirtVMInitialization(v.([]interface{}))
		if err != nil {
			return err
		}
		if initialization != nil {
			vmBuilder.Initialization(initialization)
		}
	}

	vm, err := vmBuilder.Build()
	if err != nil {
		return err
	}

	resp, err := vmsService.Add().
		Vm(vm).
		Send()
	if err != nil {
		return err
	}

	newVM, ok := resp.Vm()
	if !ok {
		d.SetId("")
		return nil
	}
	d.SetId(newVM.MustId())

	vmService := conn.SystemService().VmsService().VmService(d.Id())

	// Do attach disks
	if v, ok := d.GetOk("attached_disk"); ok {
		err = ovirtAttachDisks(v.(*schema.Set).List(), d.Id(), meta)
		if err != nil {
			return err
		}
	}

	// Do attach vnics
	if v, ok := d.GetOk("vnic"); ok {
		err = ovirtAttachVnics(v.(*schema.Set).List(), d.Id(), meta)
		if err != nil {
			return err
		}
	}

	// Try to start VM
	_, err = vmService.Start().Send()
	if err != nil {
		return err
	}
	// Wait for 5 minutes until vm is up
	err = conn.WaitForVM(d.Id(), ovirtsdk4.VMSTATUS_UP, 5*time.Minute)
	if err != nil {
		return err
	}

	return resourceOvirtVMRead(d, meta)
}

func resourceOvirtVMUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)
	vmService := conn.SystemService().VmsService().VmService(d.Id())

	paramVM := ovirtsdk4.NewVmBuilder()

	attributeUpdated := false

	d.Partial(true)

	// Allowed to attach/detach disk attachments at any VM status
	if d.HasChange("attached_disk") {
		o, n := d.GetChange("attached_disk")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		rl := os.Difference(ns).List()
		al := ns.Difference(os).List()
		if len(rl) > 0 {
			for _, v := range rl {
				vmap := v.(map[string]interface{})
				resource.Retry(2*time.Minute, func() *resource.RetryError {
					_, err := vmService.DiskAttachmentsService().
						AttachmentService(vmap["disk_attachment_id"].(string)).
						Remove().
						Send()
					if err != nil {
						if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
							// Ignore the non existed one when detaching
							return nil
						}
						return resource.RetryableError(fmt.Errorf("failed to detach disk: %s, wait for next check", err))
					}
					return nil
				})
			}
		}
		if len(al) > 0 {
			err := ovirtAttachDisks(al, d.Id(), meta)
			if err != nil {
				return err
			}
		}
		d.SetPartial("attached_disk")
	}

	if d.HasChange("vnic") {
		o, n := d.GetChange("vnic")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		rl := os.Difference(ns).List()
		al := ns.Difference(os).List()

		if len(rl) > 0 {
			err := ovirtDetachVnics(rl, d.Id(), meta)
			if err != nil {
				return err
			}
		}

		// Allowd to attached a vnic at any VM status
		if len(al) > 0 {
			err := ovirtAttachVnics(al, d.Id(), meta)
			if err != nil {
				return err
			}
		}

		d.SetPartial("vnic")
	}

	// initialization is a built-in attribute of VM that could be changed
	// at any conditions.
	if d.HasChange("initialization") {
		if v, ok := d.GetOk("initialization"); ok {
			initialization, err := expandOvirtVMInitialization(v.([]interface{}))
			if err != nil {
				return err
			}
			paramVM.Initialization(initialization)
		}
		attributeUpdated = true
	}

	if attributeUpdated {
		_, err := vmService.Update().Vm(paramVM.MustBuild()).Send()
		if err != nil {
			return err
		}
	}

	// Any way, ensure the VM is UP
	vmService.Start().Send()
	err := conn.WaitForVM(d.Id(), ovirtsdk4.VMSTATUS_UP, 2*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to start vm: %s", err)
	}

	d.Partial(false)
	return resourceOvirtVMRead(d, meta)
}

func resourceOvirtVMRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	getVmresp, err := conn.SystemService().VmsService().
		VmService(d.Id()).Get().Send()
	if err != nil {
		if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
			d.SetId("")
			return nil
		}
		return err
	}

	vm, ok := getVmresp.Vm()

	if !ok {
		d.SetId("")
		return nil
	}
	d.Set("name", vm.MustName())
	d.Set("status", vm.MustStatus())
	d.Set("cores", vm.MustCpu().MustTopology().MustCores())
	d.Set("sockets", vm.MustCpu().MustTopology().MustSockets())
	d.Set("threads", vm.MustCpu().MustTopology().MustThreads())
	d.Set("cluster_id", vm.MustCluster().MustId())

	template, err := conn.FollowLink(vm.MustTemplate())
	if err != nil {
		return err
	}
	if v, ok := template.(*ovirtsdk4.Template); ok {
		d.Set("template", v.MustName())
	}

	if v, ok := vm.Initialization(); ok {
		if err = d.Set("initialization", flattenOvirtVMInitialization(v)); err != nil {
			return fmt.Errorf("error setting initialization: %s", err)
		}
	}

	attachments, err := conn.FollowLink(vm.MustDiskAttachments())
	if err != nil {
		return err
	}
	if v, ok := attachments.(*ovirtsdk4.DiskAttachmentSlice); ok && len(v.Slice()) > 0 {
		if err = d.Set("attached_disk", flattenOvirtVMDiskAttachments(v.Slice())); err != nil {
			return fmt.Errorf("error setting attached_disk: %s", err)
		}
	}

	nicSlice, err := conn.FollowLink(vm.MustNics())
	if err != nil {
		return err
	}
	if nics, ok := nicSlice.(*ovirtsdk4.NicSlice); ok && len(nics.Slice()) > 0 {
		if err = d.Set("vnic", flattenOvirtVMVnic(nics.Slice())); err != nil {
			return fmt.Errorf("error setting vnic: %s", err)
		}
	}

	return nil
}

func resourceOvirtVMDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	vmService := conn.SystemService().VmsService().VmService(d.Id())

	return resource.Retry(3*time.Minute, func() *resource.RetryError {
		getVMResp, err := vmService.Get().Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				return nil
			}
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

		detachOnly := true
		if v, ok := vm.Template(); ok {
			t, err := conn.FollowLink(v)
			if err != nil {
				return resource.RetryableError(fmt.Errorf("Get template failed and got an error: %v", err))
			}
			if t, ok := t.(*ovirtsdk4.Template); ok {
				if t.MustName() != "Blank" {
					detachOnly = false
				}
			}
		}

		_, err = vmService.Remove().
			DetachOnly(detachOnly). // DetachOnly indicates without removing disks attachments
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				return nil
			}
			return resource.RetryableError(fmt.Errorf("Delete instance timeout and got an error: %v", err))
		}

		return nil

	})
}

func expandOvirtVMInitialization(l []interface{}) (*ovirtsdk4.Initialization, error) {
	if len(l) == 0 {
		return nil, nil
	}
	s := l[0].(map[string]interface{})
	initializationBuilder := ovirtsdk4.NewInitializationBuilder()
	if v, ok := s["host_name"]; ok {
		initializationBuilder.HostName(v.(string))
	}
	if v, ok := s["timezone"]; ok {
		initializationBuilder.Timezone(v.(string))
	}
	if v, ok := s["user_name"]; ok {
		initializationBuilder.UserName(v.(string))
	}
	if v, ok := s["custom_script"]; ok {
		initializationBuilder.CustomScript(v.(string))
	}
	if v, ok := s["authorized_ssh_key"]; ok {
		initializationBuilder.AuthorizedSshKeys(v.(string))
	}
	if v, ok := s["dns_servers"]; ok {
		initializationBuilder.DnsServers(v.(string))
	}
	if v, ok := s["dns_search"]; ok {
		initializationBuilder.DnsSearch(v.(string))
	}
	if v, ok := s["nic_configuration"]; ok {
		ncs, err := expandOvirtVMNicConfigurations(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		if len(ncs) > 0 {
			initializationBuilder.NicConfigurationsOfAny(ncs...)
		}
	}
	return initializationBuilder.Build()
}

func expandOvirtVMNicConfigurations(l []interface{}) ([]*ovirtsdk4.NicConfiguration, error) {
	nicConfs := make([]*ovirtsdk4.NicConfiguration, len(l))
	for i, v := range l {
		vmap := v.(map[string]interface{})
		ncbuilder := ovirtsdk4.NewNicConfigurationBuilder()
		ncbuilder.Name(vmap["label"].(string))
		ncbuilder.BootProtocol(ovirtsdk4.BootProtocol(vmap["boot_proto"].(string)))
		if v, ok := vmap["on_boot"]; ok {
			ncbuilder.OnBoot(v.(bool))
		}
		address, addressOK := vmap["address"]
		netmask, netmaskOK := vmap["netmask"]
		gateway, gatewayOK := vmap["gateway"]
		if addressOK || netmaskOK || gatewayOK {
			ipBuilder := ovirtsdk4.NewIpBuilder()
			if addressOK {
				ipBuilder.Address(address.(string))
			}
			if netmaskOK {
				ipBuilder.Netmask(netmask.(string))
			}
			if gatewayOK {
				ipBuilder.Gateway(gateway.(string))
			}
			ncbuilder.IpBuilder(ipBuilder)
		}
		nc, err := ncbuilder.Build()
		if err != nil {
			return nil, err
		}
		nicConfs[i] = nc
	}
	return nicConfs, nil
}

func expandOvirtVMDiskAttachment(d interface{}, disk *ovirtsdk4.Disk) (*ovirtsdk4.DiskAttachment, error) {
	dmap := d.(map[string]interface{})
	builder := ovirtsdk4.NewDiskAttachmentBuilder()
	if disk != nil {
		builder.Disk(disk)
	}
	if v, ok := dmap["interface"]; ok {
		builder.Interface(ovirtsdk4.DiskInterface(v.(string)))
	}
	if v, ok := dmap["bootable"]; ok {
		builder.Bootable(v.(bool))
	}
	if v, ok := dmap["active"]; ok {
		builder.Active(v.(bool))
	}
	if v, ok := dmap["logical_name"]; ok {
		builder.LogicalName(v.(string))
	}
	if v, ok := dmap["pass_discard"]; ok {
		builder.PassDiscard(v.(bool))
	}
	if v, ok := dmap["read_only"]; ok {
		builder.ReadOnly(v.(bool))
	}
	if v, ok := dmap["use_scsi_reservation"]; ok {
		builder.UsesScsiReservation(v.(bool))
	}

	return builder.Build()
}

func ovirtAttachDisks(s []interface{}, vmID string, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)
	vmService := conn.SystemService().VmsService().VmService(vmID)
	for _, v := range s {
		attachment := v.(map[string]interface{})
		diskService := conn.SystemService().DisksService().
			DiskService(attachment["disk_id"].(string))
		var disk *ovirtsdk4.Disk
		err := resource.Retry(30*time.Second, func() *resource.RetryError {
			getDiskResp, err := diskService.Get().Send()
			if err != nil {
				return resource.RetryableError(err)
			}
			disk = getDiskResp.MustDisk()
			if disk.MustStatus() == ovirtsdk4.DISKSTATUS_LOCKED {
				return resource.RetryableError(fmt.Errorf("disk is locked, wait for next check"))
			}
			return nil
		})
		if err != nil {
			return err
		}

		da, err := expandOvirtVMDiskAttachment(v, disk)
		if err != nil {
			return err
		}

		err = resource.Retry(2*time.Minute, func() *resource.RetryError {
			addAttachmentResp, err := vmService.DiskAttachmentsService().
				Add().
				Attachment(da).
				Send()
			if err != nil {
				return resource.RetryableError(fmt.Errorf("failed to attach disk: %s, wait for next check", err))
			}
			_, ok := addAttachmentResp.Attachment()
			if !ok {
				return resource.RetryableError(fmt.Errorf("failed to attach disk: not exists in response, wait for next check"))
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func ovirtAttachVnics(s []interface{}, vmID string, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)
	vmService := conn.SystemService().VmsService().VmService(vmID)

	for _, v := range s {
		vmap := v.(map[string]interface{})
		getResp, err := vmService.NicsService().
			Add().
			Nic(
				ovirtsdk4.NewNicBuilder().
					Name(vmap["name"].(string)).
					VnicProfile(
						ovirtsdk4.NewVnicProfileBuilder().
							Id(vmap["vnic_profile_id"].(string)).
							MustBuild()).
					MustBuild()).
			Send()
		if err != nil {
			return err
		}
		if _, ok := getResp.Nic(); !ok {
			return fmt.Errorf("failed to add nic: response not contains the nic")
		}
	}

	return nil
}

func ovirtDetachVnics(s []interface{}, vmID string, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)
	vmService := conn.SystemService().VmsService().VmService(vmID)
	getVMResp, err := vmService.Get().Send()
	if err != nil {
		return nil
	}
	templateLink, err := conn.FollowLink(getVMResp.MustVm().MustTemplate())
	if err != nil {
		return err
	}
	template, ok := templateLink.(*ovirtsdk4.Template)
	if !ok {
		return fmt.Errorf("failed to get template from vm")
	}

	for _, v := range s {
		vmap := v.(map[string]interface{})
		log.Printf("[DEBUG] Now to detach vnic: %#v\n", vmap)
		nicService := vmService.NicsService().NicService(vmap["vnic_id"].(string))
		// Template-based VM support hotunplug
		if template.MustName() != "Blank" {
			// Deactivate and remove it
			err := resource.Retry(1*time.Minute, func() *resource.RetryError {
				_, err := nicService.Deactivate().Send()
				if err != nil {
					return resource.RetryableError(err)
				}
				_, err = nicService.Remove().Send()
				if err != nil {
					return resource.RetryableError(err)
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			log.Printf("[DEBUG] Template is Blank, so Poweroff VM\n")
			// If template is Blank, remove the nic after poweroffing VM
			vmService.Stop().Send()
			err = conn.WaitForVM(vmID, ovirtsdk4.VMSTATUS_DOWN, 3*time.Minute)
			if err != nil {
				return err
			}
			_, err = nicService.Remove().Send()
			if err != nil {
				return err
			}
			log.Printf("[DEBUG] Detach a vnic: %s\n", vmap["vnic_id"])
		}
	}
	return nil
}

func flattenOvirtVMDiskAttachments(configured []*ovirtsdk4.DiskAttachment) []map[string]interface{} {
	diskAttachments := make([]map[string]interface{}, len(configured))
	for i, v := range configured {
		attrs := make(map[string]interface{})
		attrs["disk_attachment_id"] = v.MustId()
		attrs["disk_id"] = v.MustDisk().MustId()
		attrs["interface"] = v.MustInterface()

		if vi, ok := v.Active(); ok {
			attrs["active"] = vi
		}
		if vi, ok := v.Bootable(); ok {
			attrs["bootable"] = vi
		}
		if vi, ok := v.LogicalName(); ok {
			attrs["logical_name"] = vi
		}
		if vi, ok := v.PassDiscard(); ok {
			attrs["pass_discard"] = vi
		}
		if vi, ok := v.ReadOnly(); ok {
			attrs["read_only"] = vi
		}
		if vi, ok := v.UsesScsiReservation(); ok {
			attrs["use_scsi_reservation"] = vi
		}
		diskAttachments[i] = attrs
	}
	return diskAttachments
}

func flattenOvirtVMInitialization(configured *ovirtsdk4.Initialization) []map[string]interface{} {
	if configured == nil {
		initializations := make([]map[string]interface{}, 0)
		return initializations
	}
	initializations := make([]map[string]interface{}, 1)
	initialization := make(map[string]interface{})

	if v, ok := configured.HostName(); ok {
		initialization["host_name"] = v
	}
	if v, ok := configured.Timezone(); ok {
		initialization["timezone"] = v
	}
	if v, ok := configured.UserName(); ok {
		initialization["user_name"] = v
	}
	if v, ok := configured.CustomScript(); ok {
		initialization["custom_script"] = v
	}
	if v, ok := configured.DnsServers(); ok {
		initialization["dns_servers"] = v
	}
	if v, ok := configured.DnsSearch(); ok {
		initialization["dns_search"] = v
	}
	if v, ok := configured.AuthorizedSshKeys(); ok {
		initialization["authorized_ssh_key"] = v
	}
	if v, ok := configured.NicConfigurations(); ok {
		initialization["nic_configuration"] = flattenOvirtVMInitializationNicConfigurations(v.Slice())
	}
	initializations[0] = initialization
	return initializations
}

func flattenOvirtVMInitializationNicConfigurations(configured []*ovirtsdk4.NicConfiguration) []map[string]interface{} {
	ncs := make([]map[string]interface{}, len(configured))
	for i, v := range configured {
		attrs := make(map[string]interface{})
		if name, ok := v.Name(); ok {
			attrs["label"] = name
		}
		attrs["on_boot"] = v.MustOnBoot()
		attrs["boot_proto"] = v.MustBootProtocol()
		if ipAttrs, ok := v.Ip(); ok {
			if ipAddr, ok := ipAttrs.Address(); ok {
				attrs["address"] = ipAddr
			}
			if netmask, ok := ipAttrs.Netmask(); ok {
				attrs["netmask"] = netmask
			}
			if gateway, ok := ipAttrs.Gateway(); ok {
				attrs["gateway"] = gateway
			}
		}
		ncs[i] = attrs
	}
	return ncs
}

func flattenOvirtVMVnic(configured []*ovirtsdk4.Nic) []map[string]interface{} {
	vnics := make([]map[string]interface{}, len(configured))
	for i, v := range configured {
		attrs := make(map[string]interface{})
		attrs["vnic_id"] = v.MustId()
		attrs["name"] = v.MustName()
		attrs["vnic_profile_id"] = v.MustVnicProfile().MustId()
		vnics[i] = attrs
	}
	return vnics
}
