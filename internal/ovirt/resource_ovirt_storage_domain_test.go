package ovirt

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestStorageDomainResource(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	client := p.getTestHelper().GetClient().WithContext(context.Background())
	hosts, err := client.ListHosts()
	if err != nil {
		t.Fatalf("Failed to list hosts (%v)", err)
	}
	if len(hosts) == 0 {
		t.Fatalf("No hosts found")
	}
	hostID := hosts[0].ID()

	name := fmt.Sprintf("%s-%s", t.Name(), p.getTestHelper().GenerateRandomID(5))
	config := fmt.Sprintf(
		`
provider "ovirt" {
	mock = true
}

resource "ovirt_storage_domain" "test" {
	name                          = "%s"
	host_id                       = "%s"
	storage_type                  = "nfs"
	storage_address               = "nfs.example.com"
	storage_path                  = "/data/export"
	storage_format                = "v5"
	description                   = "test storage domain description"
	comment                       = "test storage domain comment"
	wipe_after_delete             = true
	critical_space_action_blocker = 5
	warning_low_space_indicator   = 10
}
`,
		name,
		hostID,
	)

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(
							"ovirt_storage_domain.test",
							"name",
							regexp.MustCompile(name),
						),
						resource.TestMatchResourceAttr(
							"ovirt_storage_domain.test",
							"host_id",
							regexp.MustCompile(string(hostID)),
						),
						resource.TestMatchResourceAttr(
							"ovirt_storage_domain.test",
							"storage_type",
							regexp.MustCompile("^nfs$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_storage_domain.test",
							"storage_address",
							regexp.MustCompile("^nfs.example.com$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_storage_domain.test",
							"storage_path",
							regexp.MustCompile("^/data/export$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_storage_domain.test",
							"storage_format",
							regexp.MustCompile("^v5$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_storage_domain.test",
							"description",
							regexp.MustCompile("^test storage domain description$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_storage_domain.test",
							"comment",
							regexp.MustCompile("^test storage domain comment$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_storage_domain.test",
							"wipe_after_delete",
							regexp.MustCompile("^true$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_storage_domain.test",
							"critical_space_action_blocker",
							regexp.MustCompile("^5$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_storage_domain.test",
							"warning_low_space_indicator",
							regexp.MustCompile("^10$"),
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
