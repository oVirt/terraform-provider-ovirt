package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclient "github.com/ovirt/go-ovirt-client/v2"
)

func TestTagAttachmentResource(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	name := fmt.Sprintf("%s-%s", t.Name(), p.getTestHelper().GenerateRandomID(5))
	clusterID := p.getTestHelper().GetClusterID()
	templateID := p.getTestHelper().GetBlankTemplateID()
	config := fmt.Sprintf(
		`
provider "ovirt" {
	mock = true
}

resource "ovirt_tag" "test" {
	name = "%s"
}

resource "ovirt_vm" "test" {
	cluster_id  = "%s"
	template_id = "%s"
    name        = "test"
}

resource "ovirt_vm_tag" "test" {
    vm_id = ovirt_vm.test.id
    tag_id = ovirt_tag.test.id
}
`,
		name,
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
						resource.TestMatchResourceAttr(
							"ovirt_vm_tag.test",
							"tag_id",
							regexp.MustCompile("^.+$"),
						),
						func(state *terraform.State) error {
							res := state.RootModule().Resources["ovirt_vm_tag.test"]
							vmID := res.Primary.Attributes["vm_id"]
							tags, err := p.getTestHelper().GetClient().ListVMTags(ovirtclient.VMID(vmID))
							if err != nil {
								return err
							}
							if len(tags) != 1 {
								return fmt.Errorf(
									"iIncorrect number of tags on VM (expected: %d, got: %d)",
									1,
									len(tags),
								)
							}
							tag := tags[0]
							tagID := res.Primary.Attributes["tag_id"]
							if tag.ID() != ovirtclient.TagID(tagID) {
								return fmt.Errorf("incorrect tag ID on VM (expected: %s, got: %s)", tagID, tag.ID())
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
