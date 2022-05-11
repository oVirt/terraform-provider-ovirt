package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclient "github.com/ovirt/go-ovirt-client"
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
	storagedomain_id = "%s"
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
						"storagedomain_id",
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
	storagedomain_id = "%s"
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
						"storagedomain_id",
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
