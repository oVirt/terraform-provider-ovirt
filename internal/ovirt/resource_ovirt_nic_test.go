package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func TestNICResource(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	clusterID := p.getTestHelper().GetClusterID()
	templateID := p.getTestHelper().GetBlankTemplateID()
	vnicProfileID := p.getTestHelper().GetVNICProfileID()

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: p.getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					`
provider "ovirt" {
	mock = true
}

resource "ovirt_vm" "test" {
	cluster_id  = "%s"
	template_id = "%s"
    name        = "test"
}

resource "ovirt_nic" "test" {
	vm_id           = ovirt_vm.test.id
	vnic_profile_id = "%s"
	name            = "eth0"
	mac             = "11:22:33:AA:BB:CC"
}
`,
					clusterID,
					templateID,
					vnicProfileID,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"ovirt_nic.test",
						"vnic_profile_id",
						regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(string(vnicProfileID)))),
					),
					resource.TestMatchResourceAttr(
						"ovirt_nic.test",
						"name",
						regexp.MustCompile("^eth0$"),
					),
					resource.TestMatchResourceAttr(
						"ovirt_nic.test",
						"mac",
						regexp.MustCompile("^11:22:33:AA:BB:CC$"),
					),
				),
			},
		},
	})
}

func TestNICResourceImport(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	client := p.getTestHelper().GetClient()
	clusterID := p.getTestHelper().GetClusterID()
	templateID := p.getTestHelper().GetBlankTemplateID()
	vnicProfileID := p.getTestHelper().GetVNICProfileID()

	config1 := fmt.Sprintf(`
provider "ovirt" {
	mock = true
}

resource "ovirt_vm" "test" {
	cluster_id  = "%s"
	template_id = "%s"
    name        = "test"
}
`, clusterID, templateID)
	config2 := fmt.Sprintf(`%s

resource "ovirt_nic" "test" {
	vm_id           = ovirt_vm.test.id
	vnic_profile_id = "%s"
	name            = "eth0"
}
`, config1, vnicProfileID)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: p.getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config1,
			},
			{
				Config:      config2,
				ImportState: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					nic, err := client.CreateNIC(
						ovirtclient.VMID(state.RootModule().Resources["ovirt_vm.test"].Primary.ID),
						vnicProfileID,
						"eth0",
						nil,
					)
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("%s/%s", nic.VMID(), nic.ID()), nil
				},
				ResourceName: "ovirt_nic.test",
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"ovirt_nic.test",
						"vnic_profile_id",
						regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(string(vnicProfileID)))),
					),
					resource.TestMatchResourceAttr(
						"ovirt_nic.test",
						"name",
						regexp.MustCompile("^eth0$"),
					),
				),
			},
		},
	})
}
