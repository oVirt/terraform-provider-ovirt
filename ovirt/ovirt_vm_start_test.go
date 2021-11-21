package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v2"
)

func TestVMStartResource(t *testing.T) {
	t.Parallel()

	p := newProvider(ovirtclientlog.NewTestLogger(t))
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

resource "ovirt_vm_start" "foo" {
	vm_id = ovirt_vm.foo.id
}
`,
		clusterID,
		templateID,
	)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: p.getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"ovirt_vm_start.foo",
						"status",
						regexp.MustCompile("up"),
					),
				),
			},
			{
				Config:  config,
				Destroy: true,
			},
		},
	})
}
