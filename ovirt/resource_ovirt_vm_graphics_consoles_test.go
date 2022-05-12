package ovirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestVMGraphicsConsoles(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	clusterID := p.getTestHelper().GetClusterID()
	templateID := p.getTestHelper().GetBlankTemplateID()
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

resource "ovirt_vm_graphics_consoles" "foo" {
	vm_id = ovirt_vm.foo.id
}
`,
		clusterID,
		templateID,
	)

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						func(state *terraform.State) error {
							res := state.RootModule().Resources["ovirt_vm_graphics_consoles.foo"]
							cons := res.Primary.Attributes["console"]
							if len(cons) != 0 {
								return fmt.Errorf(
									"there are still %d graphics consoles present after removal",
									len(cons),
								)
							}
							return nil
						},
					),
				},
				{
					Config:  config,
					Destroy: true,
				},
			},
		},
	)
}
