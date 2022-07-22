package ovirt

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclient "github.com/ovirt/go-ovirt-client/v2"
)

func TestOvirtWaitForIP(t *testing.T) {
	t.Parallel()

	_, err := os.Stat("./testimage/full.qcow")
	if err != nil {
		t.Skipf("Full test image is not available, please run go generate.")
	}

	p := newProvider(newTestLogger(t))
	helper := p.getTestHelper()
	client := helper.GetClient().WithContext(context.Background())
	clusterID := helper.GetClusterID()
	templateID := helper.GetBlankTemplateID()
	vnicProfileID := helper.GetVNICProfileID()
	storageDomainID := helper.GetStorageDomainID()
	config := fmt.Sprintf(
		`
provider "ovirt" {
	mock = true
}

resource "ovirt_vm" "foo" {
	cluster_id = "%s"
	template_id = "%s"
	name = "test"
}

resource "ovirt_nic" "foo" {
  name            = "eth0"
  vm_id           = ovirt_vm.foo.id
  vnic_profile_id = "%s"
}

resource "ovirt_disk_from_image" "foo" {
	storage_domain_id = "%s"
	format           = "cow"
    alias            = "test"
    sparse           = true
    source_file      = "./testimage/full.qcow"
}

resource "ovirt_disk_attachment" "foo" {
	vm_id          = ovirt_vm.foo.id
	disk_id        = ovirt_disk_from_image.foo.id
	disk_interface = "virtio_scsi"
}

resource "ovirt_vm_start" "foo" {
	vm_id = ovirt_vm.foo.id

	depends_on = [ovirt_nic.foo, ovirt_disk_attachment.foo] 
}

data "ovirt_wait_for_ip" "test" {
    vm_id = ovirt_vm_start.foo.vm_id
}
`,
		clusterID,
		templateID,
		vnicProfileID,
		storageDomainID,
	)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: p.getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: func(state *terraform.State) error {
					vmID := ovirtclient.VMID(state.RootModule().Resources["ovirt_vm.foo"].Primary.ID)
					vmIPs, err := client.GetVMNonLocalIPAddresses(vmID)
					if err != nil {
						return err
					}
					if len(vmIPs) == 0 {
						return fmt.Errorf("no non-local IP addresses found")
					}
					return nil
				},
			},
			{
				Config:  config,
				Destroy: true,
			},
		},
	})
}
