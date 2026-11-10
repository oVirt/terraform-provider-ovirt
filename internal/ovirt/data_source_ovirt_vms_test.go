package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestVMsDataSource(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	clusterID := p.getTestHelper().GetClusterID()
	templateID := p.getTestHelper().GetBlankTemplateID()

	vmName1 := "test-vm-search-1"
	vmName2 := "test-vm-search-2"

	config := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		resource "ovirt_vm" "vm1" {
			cluster_id  = "%s"
			template_id = "%s"
			name        = "%s"
			comment     = "First test VM"
		}

		resource "ovirt_vm" "vm2" {
			cluster_id  = "%s"
			template_id = "%s"
			name        = "%s"
			comment     = "Second test VM"
		}

		data "ovirt_vms" "by_name" {
			name = "%s"

			depends_on = [ovirt_vm.vm1, ovirt_vm.vm2]
		}

		data "ovirt_vms" "by_cluster" {
			cluster_id = "%s"

			depends_on = [ovirt_vm.vm1, ovirt_vm.vm2]
		}
	`, clusterID, templateID, vmName1, clusterID, templateID, vmName2, vmName1, clusterID)

	configNotFound := `
		provider "ovirt" {
			mock = true
		}

		data "ovirt_vms" "not_found" {
			name          = "non-existent-vm-xyz-12345"
			fail_on_empty = true
		}
	`

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.ovirt_vms.by_name", "vms.#", "1"),
						resource.TestCheckTypeSetElemNestedAttrs(
							"data.ovirt_vms.by_name",
							"vms.*",
							map[string]string{
								"name":    vmName1,
								"comment": "First test VM",
							},
						),
						resource.TestCheckResourceAttrSet("data.ovirt_vms.by_cluster", "vms.#"),
					),
				},
				{
					Config:      configNotFound,
					ExpectError: regexp.MustCompile("No virtual machines found"),
				},
			},
		},
	)
}
