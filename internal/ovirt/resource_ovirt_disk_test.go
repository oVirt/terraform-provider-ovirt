//nolint:revive
package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func TestDiskResource(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	storageDomainID := p.getTestHelper().GetStorageDomainID()

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: p.getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					`
provider "ovirt" {
	mock = true
}

resource "ovirt_disk" "foo" {
	storage_domain_id = "%s"
	format           = "raw"
    size             = 1048576
    alias            = "test"
    sparse           = true
}
`,
					storageDomainID,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"ovirt_disk.foo",
						"storage_domain_id",
						regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(string(storageDomainID)))),
					),
				),
			},
		},
	})
}

func TestDiskResourceImport(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	client := p.getTestHelper().GetClient()
	storageDomainID := p.getTestHelper().GetStorageDomainID()

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: p.getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					`
provider "ovirt" {
	mock = true
}

resource "ovirt_disk" "foo" {
	storage_domain_id = "%s"
	format           = "raw"
    size             = 1048576
    alias            = "test"
    sparse           = true
}
`,
					storageDomainID,
				),
				ResourceName: "ovirt_disk.foo",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					disk, err := client.CreateDisk(
						storageDomainID,
						ovirtclient.ImageFormatRaw,
						1048576,
						nil,
					)
					if err != nil {
						return "", fmt.Errorf("failed to create import test disk (%w)", err)
					}
					return string(disk.ID()), nil
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"ovirt_disk.foo",
						"storage_domain_id",
						regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(string(storageDomainID)))),
					),
					resource.TestMatchResourceAttr(
						"ovirt_disk.foo",
						"format",
						regexp.MustCompile(fmt.Sprintf("^%s$", ovirtclient.ImageFormatRaw)),
					),
				),
			},
		},
	})
}

func TestDiskResourceSparse(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	storageDomainID := p.getTestHelper().GetStorageDomainID()

	baseConfig := `
		provider "ovirt" {
			mock = true
		}

		resource "ovirt_disk" "foo" {
			storage_domain_id = "%s"
			format           = "raw"
			size             = 1048576
			alias            = "test"
			sparse           = %s
		}`

	testcases := []struct {
		inputSparse    string
		expectedSparse bool
	}{
		{
			inputSparse:    "null",
			expectedSparse: false,
		},
		{
			inputSparse:    "false",
			expectedSparse: false,
		},
		{
			inputSparse:    "true",
			expectedSparse: true,
		},
	}

	for _, tc := range testcases {
		tcName := fmt.Sprintf("disk resource sparse=%s", tc.inputSparse)
		t.Run(tcName, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProviderFactories: p.getProviderFactories(),
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(
							baseConfig,
							storageDomainID,
							tc.inputSparse,
						),
						Check: resource.ComposeTestCheckFunc(
							func(s *terraform.State) error {
								client := p.getTestHelper().GetClient()
								diskID := s.RootModule().Resources["ovirt_disk.foo"].Primary.ID
								disk, err := client.GetDisk(ovirtclient.DiskID(diskID))
								if err != nil {
									return err
								}

								if disk.Sparse() != tc.expectedSparse {
									return fmt.Errorf("Expected sparse to be %t, but got %t",
										tc.expectedSparse,
										disk.Sparse())
								}

								return nil
							},
						),
					},
				},
			})
		})
	}
}
