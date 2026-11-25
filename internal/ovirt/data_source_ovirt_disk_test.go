package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestDiskDataSource(t *testing.T) {
	p := newProvider(newTestLogger(t))
	helper := p.getTestHelper()
	storageDomainID := helper.GetStorageDomainID()

	config := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		resource "ovirt_disk" "test_disk" {
			storage_domain_id = "%s"
			format            = "raw"
			size              = 1048576
			alias             = "my-custom-test-disk"
		}

		data "ovirt_disk" "by_alias" {
			alias = ovirt_disk.test_disk.alias
		}

		data "ovirt_disk" "by_id" {
			id = ovirt_disk.test_disk.id
		}

		output "disk_id_by_alias" {
			value = data.ovirt_disk.by_alias.id
		}

		output "disk_alias_by_id" {
			value = data.ovirt_disk.by_id.alias
		}
	`, storageDomainID)

	configNotFound := `
		provider "ovirt" {
			mock = true
		}

		data "ovirt_disk" "test_not_found" {
			alias = "non-existent-disk-xyz-123"
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
							outputID := s.RootModule().Outputs["disk_id_by_alias"].Value.(string)
							outputAlias := s.RootModule().Outputs["disk_alias_by_id"].Value.(string)
							if outputID == "" {
								return fmt.Errorf("expected non-empty disk ID")
							}
							if outputAlias != "my-custom-test-disk" {
								return fmt.Errorf("expected disk alias 'my-custom-test-disk', got %q", outputAlias)
							}
							return nil
						},
						resource.TestCheckResourceAttr("data.ovirt_disk.by_alias", "alias", "my-custom-test-disk"),
						resource.TestCheckResourceAttr("data.ovirt_disk.by_alias", "format", "raw"),
						resource.TestCheckResourceAttr("data.ovirt_disk.by_id", "alias", "my-custom-test-disk"),
						resource.TestCheckResourceAttr("data.ovirt_disk.by_id", "format", "raw"),
					),
				},
				{
					Config:      configNotFound,
					ExpectError: regexp.MustCompile("not found"),
				},
			},
		},
	)
}
