package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDatacenterResource(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))

	name := fmt.Sprintf("%s-%s", t.Name(), p.getTestHelper().GenerateRandomID(5))
	config := fmt.Sprintf(
		`
provider "ovirt" {
	mock = true
}

resource "ovirt_datacenter" "test" {
	name        = "%s"
	description = "test datacenter description"
	comment     = "test datacenter comment"
	local       = false
	quota_mode  = "audit"
}
`,
		name,
	)

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(
							"ovirt_datacenter.test",
							"name",
							regexp.MustCompile(name),
						),
						resource.TestMatchResourceAttr(
							"ovirt_datacenter.test",
							"description",
							regexp.MustCompile("^test datacenter description$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_datacenter.test",
							"comment",
							regexp.MustCompile("^test datacenter comment$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_datacenter.test",
							"local",
							regexp.MustCompile("^false$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_datacenter.test",
							"quota_mode",
							regexp.MustCompile("^audit$"),
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
