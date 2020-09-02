package ovirt

import (
	"fmt"
	"log"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func resourceOvirtAffinityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceOvirtAffinityGroupCreate,
		Read:   resourceOvirtAffinityGroupRead,
		Update: resourceOvirtAffinityGroupUpdate,
		Delete: resourceOvirtAffinityGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "Name of the affinity group",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Description of the affinity group",
			},
			"cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Cluster ID where the affinity group is",
			},
			"vm_positive": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Description: "Positive or negative affinity",
			},
			"vm_enforcing": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     false,
				Description: "Is the policy being enforced",
			},
			"vm_list": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    false,
				Description: "List of VMs in the affinity group",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"host_positive": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Description: "Positive or negative affinity",
			},
			"host_enforcing": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     false,
				Description: "Is the policy being enforced",
			},
			"host_list": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    false,
				Description: "List of Hosts in the affinity group",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceOvirtAffinityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	agBuilder := ovirtsdk4.NewAffinityGroupBuilder()

	agBuilder.Name(d.Get("name").(string))

	if desc, ok := d.GetOk("description"); ok {
		agBuilder.Description(desc.(string))
	}

	vmRuleBuilder := ovirtsdk4.NewAffinityRuleBuilder()
	if _, ok := d.GetOk("vm_list"); ok {
		vmRuleBuilder.Enabled(true)
		vmRuleBuilder.Positive(d.Get("vm_positive").(bool))
		vmRuleBuilder.Enforcing(d.Get("vm_enforcing").(bool))
	} else {
		vmRuleBuilder.Enabled(false)
	}
	agBuilder.VmsRule(vmRuleBuilder.MustBuild())

	hostRuleBuilder := ovirtsdk4.NewAffinityRuleBuilder()
	if _, ok := d.GetOk("host_list"); ok {
		hostRuleBuilder.Enabled(true)
		hostRuleBuilder.Positive(d.Get("host_positive").(bool))
		hostRuleBuilder.Enforcing(d.Get("host_enforcing").(bool))
	} else {
		hostRuleBuilder.Enabled(false)
	}
	agBuilder.HostsRule(hostRuleBuilder.MustBuild())

	log.Printf("Creating %#v", agBuilder.MustBuild())
	addResp, err := conn.SystemService().
		ClustersService().
		ClusterService(d.Get("cluster_id").(string)).
		AffinityGroupsService().
		Add().
		Group(agBuilder.MustBuild()).
		Send()

	if err != nil {
		log.Printf("Failed to create Affinity Group")
		return err
	}

	log.Printf("Successfully created %#v", agBuilder.MustBuild().MustName())
	d.SetId(addResp.MustGroup().MustId())

	// Add VMs to affinity group
	if vmList, ok := d.GetOk("vm_list"); ok {
		vmsService := conn.SystemService().
			ClustersService().
			ClusterService(d.Get("cluster_id").(string)).
			AffinityGroupsService().
			GroupService(addResp.MustGroup().MustId()).
			VmsService()

		if err := updateVmList(vmsService, vmList.([]interface{})); err != nil {
			return err
		}
	}

	// Add hosts to affinity group
	if hostList, ok := d.GetOk("host_list"); ok {
		hostsService := conn.SystemService().
			ClustersService().
			ClusterService(d.Get("cluster_id").(string)).
			AffinityGroupsService().
			GroupService(addResp.MustGroup().MustId()).
			HostsService()

		if err := updateHostList(hostsService, hostList.([]interface{})); err != nil {
			return err
		}
	}

	return resourceOvirtAffinityGroupRead(d, meta)
}

func resourceOvirtAffinityGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	agService := conn.SystemService().
		ClustersService().
		ClusterService(d.Get("cluster_id").(string)).
		AffinityGroupsService().
		GroupService(d.Id())

	affinityGroupResp, err := agService.Get().Send()
	if err != nil {
		if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
			d.SetId("")
			return nil
		}
		return err
	}

	affinityGroup := affinityGroupResp.MustGroup()

	d.Set("name", affinityGroup.MustName())
	if desc, ok := affinityGroup.Description(); ok {
		d.Set("description", desc)
	}
	d.Set("host_enabled", affinityGroup.MustHostsRule().MustEnabled())
	d.Set("host_enforcing", affinityGroup.MustHostsRule().MustEnforcing())
	d.Set("host_positive", affinityGroup.MustHostsRule().MustPositive())
	d.Set("vm_enabled", affinityGroup.MustVmsRule().MustEnabled())
	d.Set("vm_enforcing", affinityGroup.MustVmsRule().MustEnforcing())
	d.Set("vm_positive", affinityGroup.MustVmsRule().MustPositive())

	hosts := affinityGroup.MustHosts().Slice()
	hostNames := make([]string, len(hosts))
	for i, h := range hosts {
		hostNames[i] = h.MustId()
	}
	sort.Strings(hostNames)
	d.Set("host_list", hostNames)

	vms := affinityGroup.MustVms().Slice()
	vmNames := make([]string, len(vms))
	for i, v := range vms {
		vmNames[i] = v.MustId()
	}
	sort.Strings(vmNames)
	d.Set("vm_list", vmNames)

	return nil
}

func resourceOvirtAffinityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	group := ovirtsdk4.NewAffinityGroupBuilder()
	attributeUpdate := false

	if d.HasChange("name") {
		group.Name(d.Get("name").(string))
		attributeUpdate = true
	}
	if d.HasChange("description") {
		group.Description(d.Get("description").(string))
		attributeUpdate = true
	}
	if d.HasChange("cluster_id") {
		group.Cluster(
			ovirtsdk4.NewClusterBuilder().
				Id(d.Get("cluster_id").(string)).
				MustBuild())
		attributeUpdate = true
	}

	vmRuleBuilder := ovirtsdk4.NewAffinityRuleBuilder()
	vmRuleUpdate := false
	if d.HasChange("vm_positive") {
		vmRuleBuilder.Positive(d.Get("vm_positive").(bool))
		vmRuleUpdate = true
	}
	if d.HasChange("vm_enforcing") {
		vmRuleBuilder.Enforcing(d.Get("vm_enforcing").(bool))
		vmRuleUpdate = true
	}
	if d.HasChange("vm_list") {
		vmRuleBuilder.Enabled(len(d.Get("vm_list").([]interface{})) > 0)
		vmRuleUpdate = true
	}
	if vmRuleUpdate {
		group.VmsRule(vmRuleBuilder.MustBuild())
		attributeUpdate = true
	}

	hostRuleBuilder := ovirtsdk4.NewAffinityRuleBuilder()
	hostRuleUpdate := false
	if d.HasChange("host_positive") {
		hostRuleBuilder.Positive(d.Get("host_positive").(bool))
		hostRuleUpdate = true
	}
	if d.HasChange("host_enforcing") {
		hostRuleBuilder.Enforcing(d.Get("host_enforcing").(bool))
		hostRuleUpdate = true
	}
	if d.HasChange("host_list") {
		hostRuleBuilder.Enabled(len(d.Get("host_list").([]interface{})) > 0)
		hostRuleUpdate = true
	}
	if hostRuleUpdate {
		group.HostsRule(hostRuleBuilder.MustBuild())
		attributeUpdate = true
	}

	if attributeUpdate {
		log.Printf("[DEBUG] Updating %#v", group.MustBuild())
		_, err := conn.SystemService().
			ClustersService().
			ClusterService(d.Get("cluster_id").(string)).
			AffinityGroupsService().
			GroupService(d.Id()).
			Update().
			Group(group.MustBuild()).
			Send()
		if err != nil {
			log.Printf("[DEBUG] Error updating affinity group (%s): %s", d.Id(), err)
			return err
		}
	}

	if d.HasChange("vm_list") {
		vmsService := conn.SystemService().
			ClustersService().
			ClusterService(d.Get("cluster_id").(string)).
			AffinityGroupsService().
			GroupService(d.Id()).
			VmsService()

		if err := updateVmList(vmsService, d.Get("vm_list").([]interface{})); err != nil {
			return err
		}
	}

	if d.HasChange("host_list") {
		hostsService := conn.SystemService().
			ClustersService().
			ClusterService(d.Get("cluster_id").(string)).
			AffinityGroupsService().
			GroupService(d.Id()).
			HostsService()

		if err := updateHostList(hostsService, d.Get("host_list").([]interface{})); err != nil {
			return err
		}
	}

	return resourceOvirtAffinityGroupRead(d, meta)
}

func resourceOvirtAffinityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)
	_, err := conn.SystemService().
		ClustersService().
		ClusterService(d.Get("cluster_id").(string)).
		AffinityGroupsService().
		GroupService(d.Id()).
		Remove().
		Send()
	if err != nil {
		if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
			return nil
		}
		return err
	}
	return nil
}

func stringInSlice(a string, list []interface{}) bool {
	for _, b := range list {
		if b.(string) == a {
			return true
		}
	}
	return false
}

func updateVmList(vmsService *ovirtsdk4.AffinityGroupVmsService, vmList []interface{}) error {
	// Add VMs to affinity group
	if vms, ok := vmsService.List().MustSend().Vms(); ok {
		currentVms := vms.Slice()

		// Basically implement set subtraction on both sides
		for _, v := range currentVms {
			if !stringInSlice(v.MustId(), vmList) {
				log.Printf("[DEBUG] Removing VM %s from affinity group", v.MustId())
				if _, err := vmsService.VmService(v.MustId()).Remove().Send(); err != nil {
					return err
				}
			}
		}

		for _, v := range vmList {
			exists := false
			for _, vm := range currentVms {
				if vm.MustId() == v {
					exists = true
				}
			}
			if !exists {
				log.Printf("[DEBUG] Adding VM  %s to affinity group", v.(string))
				vm := ovirtsdk4.NewVmBuilder().Id(v.(string)).MustBuild()
				if _, err := vmsService.Add().Vm(vm).Send(); err != nil {
					if _, ok := err.(ovirtsdk4.XMLTagNotMatchError); !ok {
						log.Printf("[DEBUG] Failed to add vm %s to affinity group", vm.MustId())
						return err
					}
				}
			}
		}
	} else {
		return fmt.Errorf("could not get list of VMs to update")
	}
	return nil
}

func updateHostList(hostsService *ovirtsdk4.AffinityGroupHostsService, hostList []interface{}) error {
	// Add Hosts to affinity group
	if hosts, ok := hostsService.List().MustSend().Hosts(); ok {
		currentHosts := hosts.Slice()

		// Basically implement set subtraction on both sides
		for _, v := range currentHosts {
			if !stringInSlice(v.MustId(), hostList) {
				log.Printf("[DEBUG] Removing host %s from affinity group", v.MustId())
				if _, err := hostsService.HostService(v.MustId()).Remove().Send(); err != nil {
					return err
				}
			}
		}

		for _, v := range hostList {
			exists := false
			for _, host := range currentHosts {
				if host.MustId() == v {
					exists = true
				}
			}
			if !exists {
				log.Printf("[DEBUG] Adding host  %s to affinity group", v.(string))
				host := ovirtsdk4.NewHostBuilder().Id(v.(string)).MustBuild()
				if _, err := hostsService.Add().Host(host).Send(); err != nil {
					if _, ok := err.(ovirtsdk4.XMLTagNotMatchError); !ok {
						log.Printf("[DEBUG] Failed to add host %s to affinity group", host.MustId())
						return err
					}
				}
			}
		}
	} else {
		return fmt.Errorf("could not get list of hosts to update")
	}
	return nil
}
