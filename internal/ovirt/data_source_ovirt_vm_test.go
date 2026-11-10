package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestVMDataSource(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	helper := p.getTestHelper()
	clusterID := helper.GetClusterID()
	templateID := helper.GetBlankTemplateID()
	vmName := helper.GenerateTestResourceName(t)

	config := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		resource "ovirt_vm" "test_vm" {
			cluster_id  = "%s"
			template_id = "%s"
			name        = "%s"
		}

		data "ovirt_vm" "by_name" {
			name = ovirt_vm.test_vm.name
		}

		data "ovirt_vm" "by_id" {
			id = ovirt_vm.test_vm.id
		}

		output "vm_id_by_name" {
			value = data.ovirt_vm.by_name.id
		}

		output "vm_name_by_id" {
			value = data.ovirt_vm.by_id.name
		}
	`, clusterID, templateID, vmName)

	configNotFound := `
		provider "ovirt" {
			mock = true
		}

		data "ovirt_vm" "test_not_found" {
			name = "non-existent-vm-xyz-123"
		}
	`

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							outputID := s.RootModule().Outputs["vm_id_by_name"].Value.(string)
							outputName := s.RootModule().Outputs["vm_name_by_id"].Value.(string)
							if outputID == "" {
								return fmt.Errorf("expected non-empty VM ID")
							}
							if outputName != vmName {
								return fmt.Errorf("expected VM name %q, got %q", vmName, outputName)
							}
							return nil
						},
						resource.TestCheckResourceAttr("data.ovirt_vm.by_name", "name", vmName),
						resource.TestCheckResourceAttr("data.ovirt_vm.by_name", "cluster_id", string(clusterID)),
						resource.TestCheckResourceAttr("data.ovirt_vm.by_name", "template_id", string(templateID)),
						resource.TestCheckResourceAttr("data.ovirt_vm.by_id", "name", vmName),
					),
				},
				{
					Config:      configNotFound,
					ExpectError: regexp.MustCompile("not found|Failed to get VM"),
				},
			},
		},
	)
}
