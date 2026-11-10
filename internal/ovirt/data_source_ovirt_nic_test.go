package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestNICDataSource(t *testing.T) {
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

		resource "ovirt_nic" "test" {
			vm_id           = ovirt_vm.test.id
			vnic_profile_id = "%s"
			name            = "eth0"
			mac             = "11:22:33:AA:BB:CC"
		}

		data "ovirt_nic" "by_name" {
			vm_id = ovirt_vm.test.id
			name  = ovirt_nic.test.name
		}

		data "ovirt_nic" "by_id" {
			vm_id = ovirt_vm.test.id
			id    = ovirt_nic.test.id
		}

		output "nic_id_by_name" {
			value = data.ovirt_nic.by_name.id
		}

		output "nic_name_by_id" {
			value = data.ovirt_nic.by_id.name
		}
	`, clusterID, templateID, vnicProfileID)

	configNotFound := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		resource "ovirt_vm" "test_nf" {
			cluster_id  = "%s"
			template_id = "%s"
			name        = "test_vm_nf"
		}

		data "ovirt_nic" "test_not_found" {
			vm_id = ovirt_vm.test_nf.id
			name  = "non-existent-nic"
		}
	`, clusterID, templateID)

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							outputID := s.RootModule().Outputs["nic_id_by_name"].Value.(string)
							outputName := s.RootModule().Outputs["nic_name_by_id"].Value.(string)
							if outputID == "" {
								return fmt.Errorf("expected non-empty NIC ID")
							}
							if outputName != "eth0" {
								return fmt.Errorf("expected NIC name 'eth0', got %q", outputName)
							}
							return nil
						},
						resource.TestCheckResourceAttr("data.ovirt_nic.by_name", "name", "eth0"),
						resource.TestCheckResourceAttr("data.ovirt_nic.by_name", "vnic_profile_id", string(vnicProfileID)),
						resource.TestCheckResourceAttr("data.ovirt_nic.by_name", "mac", "11:22:33:AA:BB:CC"),
						resource.TestCheckResourceAttr("data.ovirt_nic.by_id", "name", "eth0"),
						resource.TestCheckResourceAttr("data.ovirt_nic.by_id", "mac", "11:22:33:AA:BB:CC"),
					),
				},
				{
					Config:      configNotFound,
					ExpectError: regexp.MustCompile("not found|Failed to get NIC"),
				},
			},
		},
	)
}
