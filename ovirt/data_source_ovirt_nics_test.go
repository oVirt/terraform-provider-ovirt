package ovirt_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccOvirtNicsDataSource_nameRegexFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	id := suite.GenerateRandomID(5)
	diskName := fmt.Sprintf("terraform-test-%s-disk", id)
	vmName := fmt.Sprintf("terraform-test-%s-vm", id)
	netName := fmt.Sprintf("terraform-test-%s-net", id)
	vnicProfileName := fmt.Sprintf("terraform-test-%s-vnic-profile", id)
	nicContext, err := suite.CreateNicContext(netName, vnicProfileName)
	if nicContext != nil {
		defer func() {
			if err := suite.DestroyNicContext(nicContext); err != nil {
				t.Fatal(err)
			}
		}()
	}
	if err != nil {
		t.Fatal(err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "ovirt_image_transfer" "disk" {
  alias = "%s"
  source_url = "%s"
  storage_domain_id = "%s"
  sparse = true
}

resource "ovirt_vm" "vm" {
  name        = "%s"
  cluster_id  = "%s"
  template_id = "%s"
  auto_start  = false

  os {
    type = "other"
  }

  block_device {
    interface = "virtio"
    disk_id   = ovirt_image_transfer.disk.disk_id
    size      = 1
  }
}

resource "ovirt_vnic" "nic1" {
  name            = "eth0"
  vm_id           = ovirt_vm.vm.id
  vnic_profile_id = "%s"
}

data "ovirt_nics" "name_regex_filtered_nics" {
  name_regex = "^${ovirt_vnic.nic1.name}$"
  vm_id = ovirt_vm.vm.id
}
`,
					diskName,
					fmt.Sprintf("file://%s", strings.ReplaceAll(suite.TestImageSourcePath(), "\\", "/")),
					suite.StorageDomainID(),
					vmName,
					suite.ClusterID(),
					suite.BlankTemplateID(),
					nicContext.VnicProfile.MustId(),
				),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource(
						"data.ovirt_nics.name_regex_filtered_nics",
					),
					resource.TestCheckResourceAttr(
						"data.ovirt_nics.name_regex_filtered_nics",
						"nics.#",
						"1",
					),
					resource.TestMatchResourceAttr(
						"data.ovirt_nics.name_regex_filtered_nics",
						"nics.0.name",
						regexp.MustCompile("^eth0$")),
				),
			},
		},
	})
}
