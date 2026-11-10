package ovirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestNICsDataSource(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	clusterID := p.getTestHelper().GetClusterID()
	templateID := p.getTestHelper().GetBlankTemplateID()
	vnicProfileID := p.getTestHelper().GetVNICProfileID()

	config := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		resource "ovirt_vm" "test" {
			cluster_id  = "%s"
			template_id = "%s"
			name        = "test_vm"
		}

		resource "ovirt_nic" "nic1" {
			vm_id           = ovirt_vm.test.id
			vnic_profile_id = "%s"
			name            = "eth0"
			mac             = "11:22:33:AA:BB:CC"
		}

		resource "ovirt_nic" "nic2" {
			vm_id           = ovirt_vm.test.id
			vnic_profile_id = "%s"
			name            = "eth1"
		}

		data "ovirt_nics" "all" {
			vm_id = ovirt_vm.test.id

			depends_on = [ovirt_nic.nic1, ovirt_nic.nic2]
		}
	`, clusterID, templateID, vnicProfileID, vnicProfileID)

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.ovirt_nics.all", "nics.#", "2"),
					),
				},
			},
		},
	)
}
