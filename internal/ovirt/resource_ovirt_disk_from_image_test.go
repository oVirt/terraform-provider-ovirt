package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestImageUpload(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	storageDomainID := p.getTestHelper().GetStorageDomainID()

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(
						`
provider "ovirt" {
	mock = true
}

resource "ovirt_disk_from_image" "foo" {
	storage_domain_id = "%s"
	format           = "raw"
    alias            = "test"
    sparse           = true
    source_file      = "./testimage/image"
}
`,
						storageDomainID,
					),
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(
							"ovirt_disk_from_image.foo",
							"storage_domain_id",
							regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(string(storageDomainID)))),
						),
						resource.TestMatchResourceAttr(
							"ovirt_disk_from_image.foo",
							"size",
							regexp.MustCompile(fmt.Sprintf("^%d$", 1024*1024)),
						),
					),
				},
			},
		},
	)
}
