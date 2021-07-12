package ovirt_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	govirt "github.com/ovirt/go-ovirt-client"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

//TODO fix this test
func DisabledTestAccOvirtAffinityGroup_basic(t *testing.T) {
	var affinityGroup ovirtsdk4.AffinityGroup
	resourceName := "ovirt_affinity_group.affinity_group"
	rString := "testAccOvirtAffinityGroupBasic"
	cString := "Default"
	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: resourceName,
		CheckDestroy:  testAccCheckAffinityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAffinityGroupBasic(cString, rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAffinityGroupExists(resourceName, &affinityGroup),
					testAccCheckAffinityGroupBasicValues(&affinityGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rString),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Desc of affinity group %s", rString)),
					resource.TestCheckResourceAttr(resourceName, "priority", "3"),
					resource.TestCheckResourceAttr(resourceName, "host_enforcing", "false"),
					resource.TestCheckResourceAttr(resourceName, "host_positive", "false"),
					resource.TestCheckResourceAttr(resourceName, "vm_enforcing", "false"),
					resource.TestCheckResourceAttr(resourceName, "vm_positive", "true"),
					resource.TestCheckNoResourceAttr(resourceName, "host_list"),
				),
			},
			{
				Config: testAccAffinityGroupBasicUpdate(cString, rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAffinityGroupExists(resourceName, &affinityGroup),
					testAccCheckAffinityGroupBasicUpdateValues(&affinityGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rString),
					resource.TestCheckResourceAttr(resourceName, "host_enforcing", "true"),
					resource.TestCheckResourceAttr(resourceName, "host_positive", "true"),
					resource.TestCheckResourceAttr(resourceName, "vm_enforcing", "false"),
					resource.TestCheckResourceAttr(resourceName, "vm_positive", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "vm_list"),
				),
			},
		},
	})
}

func testAccCheckAffinityGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_affinity_group" {
			continue
		}
		getResp, err := conn.SystemService().
			ClustersService().
			ClusterService(rs.Primary.Attributes["cluster_id"]).
			AffinityGroupsService().
			GroupService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.Group(); ok {
			return fmt.Errorf("Affinity Group %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckAffinityGroupExists(n string, v *ovirtsdk4.AffinityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cluster ID is set")
		}
		conn := testAccProvider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()
		getResp, err := conn.SystemService().ClustersService().
			ClusterService(rs.Primary.Attributes["cluster_id"]).
			AffinityGroupsService().
			GroupService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return err
		}
		group, ok := getResp.Group()
		if ok {
			*v = *group
			return nil
		}
		return fmt.Errorf("Affinity Group %s not exist", rs.Primary.ID)
	}
}

func testAccCheckAffinityGroupBasicValues(group *ovirtsdk4.AffinityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if group.MustVmsRule().MustEnabled() != true {
			return fmt.Errorf("bad upstream state, VM rule not enabled")
		}
		if group.MustVmsRule().MustEnforcing() != false {
			return fmt.Errorf("bad upstream state, VM rule enforcing")
		}
		if group.MustVmsRule().MustPositive() != true {
			return fmt.Errorf("bad upstream state, VM rule not positive affinity")
		}
		if group.MustHostsRule().MustEnabled() != false {
			return fmt.Errorf("bad upstream state, Hosts rule enabled")
		}
		return nil
	}
}

func testAccCheckAffinityGroupBasicUpdateValues(group *ovirtsdk4.AffinityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if group.MustHostsRule().MustEnabled() != true {
			return fmt.Errorf("bad upstream state, host rule not enabled")
		}
		if group.MustHostsRule().MustEnforcing() != true {
			return fmt.Errorf("bad upstream state, host rule not enforcing")
		}
		if group.MustHostsRule().MustPositive() != true {
			return fmt.Errorf("bad upstream state, host rule not positive affinity")
		}
		if group.MustVmsRule().MustEnabled() != false {
			return fmt.Errorf("bad upstream state, vms rule enabled")
		}
		return nil
	}
}

func testAccAffinityGroupClusterDef(cluster string) string {
	return fmt.Sprintf(`
data "ovirt_clusters" "c" {
  search = {
    criteria       = "name = %s"
    case_sensitive = false
  }
}

data "ovirt_hosts" "h" {
  search = {
    criteria       = "cluster = %s"
    case_sensitive = false
  }
}

data "ovirt_vms" "v" {
  search = {
    criteria       = "cluster = %s"
    case_sensitive = false
  }
}

locals {
  hosts         = sort([for h in data.ovirt_hosts.h.hosts : h.id])
  vms           = sort([for v in data.ovirt_vms.v.vms : v.id])
  cluster       = data.ovirt_clusters.c.clusters.0
  datacenter_id = local.cluster.datacenter_id
}
`, cluster, cluster, cluster)
}

func testAccAffinityGroupBasic(cString string, rString string) string {
	return testAccAffinityGroupClusterDef(cString) + fmt.Sprintf(`
resource "ovirt_affinity_group" "affinity_group" {
  name = "%s"
  description = "Desc of affinity group %s"
  priority = 3
  cluster_id = local.cluster.id

  vm_enforcing = false
  vm_positive = true
  vm_list = local.vms

}
`, rString, rString)
}

func testAccAffinityGroupBasicUpdate(cString string, rString string) string {
	return testAccAffinityGroupClusterDef(cString) + fmt.Sprintf(`
resource "ovirt_affinity_group" "affinity_group" {
  name = "%s"
  description = "Desc of affinity group %s"
  cluster_id = local.cluster.id

  host_enforcing = true
  host_positive = true
  host_list = local.hosts
}
`, rString, rString)
}
