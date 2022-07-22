package ovirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclient "github.com/ovirt/go-ovirt-client/v2"
)

func TestDiskResizeResource(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))

	helper := p.getTestHelper()
	disk, err := helper.GetClient().CreateDisk(
		helper.GetStorageDomainID(), ovirtclient.ImageFormatRaw, 1048576,
		ovirtclient.CreateDiskParams().MustWithAlias("TestDisk").MustWithSparse(true))
	if err != nil {
		t.Fatal(err)
	}

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: p.getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					provider "ovirt" {
						mock = true
					}
					resource "ovirt_disk_resize" "resized" {
						disk_id = "%s"
						size = 2*1048576
					}`,
					disk.ID(),
				),
				Check: resource.ComposeTestCheckFunc(
					func(state *terraform.State) error {
						diskID := state.RootModule().Resources["ovirt_disk_resize.resized"].Primary.ID
						disk, err := p.getTestHelper().GetClient().GetDisk(ovirtclient.DiskID(diskID))
						if err != nil {
							return err
						}
						if disk.ProvisionedSize() != 2097152 {
							return fmt.Errorf("Expected disk size to be 2097152, but got %d", disk.ProvisionedSize())
						}
						return nil
					},
				),
			},
		},
	})
}
