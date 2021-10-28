package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v2"
)

func TestVMResource(t *testing.T) {
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
						"ovirt_vm.foo",
						"cluster_id",
						regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(clusterID))),
					),
					resource.TestMatchResourceAttr(
						"ovirt_vm.foo",
						"template_id",
						regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(templateID))),
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

func TestVMResourceImport(t *testing.T) {
	t.Parallel()

	p := newProvider(ovirtclientlog.NewTestLogger(t))
	client := p.getTestHelper().GetClient()
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
}
`,
		clusterID,
		templateID,
	)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: p.getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:       config,
				ImportState:  true,
				ResourceName: "ovirt_vm.foo",
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					vm, err := client.CreateVM(
						clusterID,
						templateID,
						nil,
					)
					if err != nil {
						return "", fmt.Errorf("failed to create test VM (%w)", err)
					}
					return vm.ID(), nil
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"ovirt_vm.foo",
						"cluster_id",
						regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(clusterID))),
					),
					resource.TestMatchResourceAttr(
						"ovirt_vm.foo",
						"template_id",
						regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(templateID))),
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

type testVM struct {
	id         string
	name       string
	comment    string
	clusterID  string
	templateID string
	status     ovirtclient.VMStatus
}

func (t *testVM) ID() string {
	return t.id
}

func (t *testVM) Name() string {
	return t.name
}

func (t *testVM) Comment() string {
	return t.comment
}

func (t *testVM) ClusterID() string {
	return t.clusterID
}

func (t *testVM) TemplateID() string {
	return t.templateID
}

func (t *testVM) Status() ovirtclient.VMStatus {
	return t.status
}

func TestVMResourceUpdate(t *testing.T) {
	t.Parallel()

	vm := &testVM{
		id:         "asdf",
		name:       "test VM",
		comment:    "This is a test VM.",
		clusterID:  "cluster-1",
		templateID: "template-1",
		status:     ovirtclient.VMStatusUp,
	}
	resourceData := schema.TestResourceDataRaw(t, vmSchema, map[string]interface{}{})
	diags := vmResourceUpdate(vm, resourceData)
	if len(diags) != 0 {
		t.Fatalf("failed to convert VM resource (%v)", diags)
	}
	compareResource(t, resourceData, "id", vm.id)
	compareResource(t, resourceData, "name", vm.name)
	compareResource(t, resourceData, "cluster_id", vm.clusterID)
	compareResource(t, resourceData, "template_id", vm.templateID)
	compareResource(t, resourceData, "status", string(vm.status))
}

func compareResource(t *testing.T, data *schema.ResourceData, field string, value string) {
	if resourceValue := data.Get(field); resourceValue != value {
		t.Fatalf("invalid resource %s: %s, expected: %s", field, resourceValue, value)
	}
}
