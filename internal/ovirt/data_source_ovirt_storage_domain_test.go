package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestStorageDomainDataSource(t *testing.T) {
	p := newProvider(newTestLogger(t))
	helper := p.getTestHelper()

	sd, err := helper.GetClient().GetStorageDomain(helper.GetStorageDomainID())
	if err != nil {
		t.Fatalf("failed to get storage domain from test helper: %v", err)
	}

	configByName := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		data "ovirt_storage_domain" "test_by_name" {
			name = "%s"
		}

		output "storage_domain_id" {
			value = data.ovirt_storage_domain.test_by_name.id
		}

		output "storage_domain_status" {
			value = data.ovirt_storage_domain.test_by_name.status
		}
	`, sd.Name())

	configByID := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}

		data "ovirt_storage_domain" "test_by_id" {
			id = "%s"
		}

		output "storage_domain_name" {
			value = data.ovirt_storage_domain.test_by_id.name
		}
	`, sd.ID())

	configNotFound := `
		provider "ovirt" {
			mock = true
		}

		data "ovirt_storage_domain" "test_not_found" {
			name = "non-existent-storage-domain"
		}
	`

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: configByName,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							outputID := s.RootModule().Outputs["storage_domain_id"].Value.(string)
							if outputID == "" {
								return fmt.Errorf("expected non-empty storage domain ID")
							}
							outputStatus := s.RootModule().Outputs["storage_domain_status"].Value.(string)
							if outputStatus != string(sd.Status()) {
								return fmt.Errorf("expected storage domain status %q, got %q", sd.Status(), outputStatus)
							}
							return nil
						},
						resource.TestCheckResourceAttr("data.ovirt_storage_domain.test_by_name", "name", sd.Name()),
					),
				},
				{
					Config: configByID,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							outputName := s.RootModule().Outputs["storage_domain_name"].Value.(string)
							if outputName != sd.Name() {
								return fmt.Errorf("expected storage domain name %q, got %q", sd.Name(), outputName)
							}
							return nil
						},
						resource.TestCheckResourceAttr("data.ovirt_storage_domain.test_by_id", "id", string(sd.ID())),
						resource.TestCheckResourceAttr("data.ovirt_storage_domain.test_by_id", "name", sd.Name()),
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
