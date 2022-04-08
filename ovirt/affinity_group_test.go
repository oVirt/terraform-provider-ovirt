package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v2"
)

func TestAffinityGroupResource(t *testing.T) {
	t.Parallel()

	p := newProvider(ovirtclientlog.NewTestLogger(t))
	clusterID := p.getTestHelper().GetClusterID()
	name := t.Name() + "_" + p.getTestHelper().GenerateRandomID(5)

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

resource "ovirt_affinity_group" "test" {
    cluster_id = "%s"
    name = "%s"
}
`,
						clusterID,
						name,
					),
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(
							"ovirt_affinity_group.test",
							"name",
							regexp.MustCompile(name),
						),
						resource.TestMatchResourceAttr(
							"ovirt_affinity_group.test",
							"hosts_rule.#",
							regexp.MustCompile("0"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_affinity_group.test",
							"vms_rule.#",
							regexp.MustCompile("0"),
						),
					),
				},
			},
		},
	)
}

func TestAffinityGroupResourceHostsRule(t *testing.T) {
	t.Parallel()

	p := newProvider(ovirtclientlog.NewTestLogger(t))
	clusterID := p.getTestHelper().GetClusterID()
	name := t.Name() + "_" + p.getTestHelper().GenerateRandomID(5)

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

resource "ovirt_affinity_group" "test" {
    cluster_id = "%s"
    name = "%s"

    hosts_rule {
        affinity  = "positive"
        enforcing = true
    }
    vms_rule {
        affinity  = "positive"
        enforcing = true
    }
}
`,
						clusterID,
						name,
					),
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(
							"ovirt_affinity_group.test",
							"hosts_rule.#",
							regexp.MustCompile("1"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_affinity_group.test",
							"vms_rule.#",
							regexp.MustCompile("1"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_affinity_group.test",
							"hosts_rule.0.affinity",
							regexp.MustCompile("positive"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_affinity_group.test",
							"hosts_rule.0.enforcing",
							regexp.MustCompile("true"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_affinity_group.test",
							"vms_rule.0.affinity",
							regexp.MustCompile("positive"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_affinity_group.test",
							"vms_rule.0.enforcing",
							regexp.MustCompile("true"),
						),
					),
				},
			},
		},
	)
}
