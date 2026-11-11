package ovirt

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestNetworkResource(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	client := p.getTestHelper().GetClient().WithContext(context.Background())
	datacenters, err := client.ListDatacenters()
	if err != nil {
		t.Fatalf("Failed to list datacenters (%v)", err)
	}
	if len(datacenters) == 0 {
		t.Fatalf("No datacenters found")
	}
	dcID := datacenters[0].ID()

	name := fmt.Sprintf("%s-%s", t.Name(), p.getTestHelper().GenerateRandomID(5))
	config := fmt.Sprintf(
		`
provider "ovirt" {
	mock = true
}

resource "ovirt_network" "foo" {
	name          = "%s"
	datacenter_id = "%s"
	description   = "test description"
	comment       = "test comment"
	vlan          = 100
}
`,
		name,
		dcID,
	)

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(
							"ovirt_network.foo",
							"name",
							regexp.MustCompile(name),
						),
						resource.TestMatchResourceAttr(
							"ovirt_network.foo",
							"datacenter_id",
							regexp.MustCompile(string(dcID)),
						),
						resource.TestMatchResourceAttr(
							"ovirt_network.foo",
							"description",
							regexp.MustCompile("^test description$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_network.foo",
							"comment",
							regexp.MustCompile("^test comment$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_network.foo",
							"vlan",
							regexp.MustCompile("^100$"),
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
