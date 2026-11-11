package ovirt

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestClusterResource(t *testing.T) {
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

resource "ovirt_cluster" "test" {
	name                              = "%s"
	datacenter_id                     = "%s"
	cpu_type                          = "Intel SandyBridge Family"
	cpu_arch                          = "x86_64"
	description                       = "test cluster description"
	comment                           = "test cluster comment"
	ballooning                        = true
	memory_policy_over_commit_percent = 150
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
							"ovirt_cluster.test",
							"name",
							regexp.MustCompile(name),
						),
						resource.TestMatchResourceAttr(
							"ovirt_cluster.test",
							"datacenter_id",
							regexp.MustCompile(string(dcID)),
						),
						resource.TestMatchResourceAttr(
							"ovirt_cluster.test",
							"cpu_type",
							regexp.MustCompile("^Intel SandyBridge Family$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_cluster.test",
							"cpu_arch",
							regexp.MustCompile("^x86_64$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_cluster.test",
							"description",
							regexp.MustCompile("^test cluster description$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_cluster.test",
							"comment",
							regexp.MustCompile("^test cluster comment$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_cluster.test",
							"ballooning",
							regexp.MustCompile("^true$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_cluster.test",
							"memory_policy_over_commit_percent",
							regexp.MustCompile("^150$"),
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
