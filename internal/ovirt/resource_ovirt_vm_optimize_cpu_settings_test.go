package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestVMOptimizeCPUSettings(t *testing.T) {
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

resource "ovirt_vm_optimize_cpu_settings" "foo" {
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
						resource.TestMatchResourceAttr(
							"ovirt_vm_optimize_cpu_settings.foo",
							"vm_id",
							regexp.MustCompile("^.+$"),
						),
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
