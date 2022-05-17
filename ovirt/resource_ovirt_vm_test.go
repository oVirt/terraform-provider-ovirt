package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v3"
)

func TestVMResource(t *testing.T) {
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
	cluster_id  = "%s"
	template_id = "%s"
    name        = "test"
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
							"ovirt_vm.foo",
							"cluster_id",
							regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(string(clusterID)))),
						),
						resource.TestMatchResourceAttr(
							"ovirt_vm.foo",
							"template_id",
							regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(string(templateID)))),
						),
						resource.TestMatchResourceAttr(
							"ovirt_vm.foo",
							"name",
							regexp.MustCompile("^test$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_vm.foo",
							"os_type",
							regexp.MustCompile("^$"),
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

func TestVMResourceImport(t *testing.T) {
	t.Parallel()

	// Special case: we are using the ovirtclientlog.NewTestLogger here because we call the client methods outside of
	// the Terraform context.
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
    name = "test"
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
					Config:       config,
					ImportState:  true,
					ResourceName: "ovirt_vm.foo",
					ImportStateIdFunc: func(state *terraform.State) (string, error) {
						vm, err := client.CreateVM(
							clusterID,
							templateID,
							"test",
							nil,
						)
						if err != nil {
							return "", fmt.Errorf("failed to create test VM (%w)", err)
						}
						return string(vm.ID()), nil
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(
							"ovirt_vm.foo",
							"cluster_id",
							regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(string(clusterID)))),
						),
						resource.TestMatchResourceAttr(
							"ovirt_vm.foo",
							"template_id",
							regexp.MustCompile(fmt.Sprintf("^%s$", regexp.QuoteMeta(string(templateID)))),
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

func TestVMResourceOSType(t *testing.T) {
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
	cluster_id  = "%s"
	template_id = "%s"
    name        = "test"
    os_type     = "rhcos_x64"
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
							"ovirt_vm.foo",
							"os_type",
							regexp.MustCompile("^rhcos_x64$"),
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

type testVM struct {
	id         ovirtclient.VMID
	name       string
	comment    string
	clusterID  ovirtclient.ClusterID
	templateID ovirtclient.TemplateID
	status     ovirtclient.VMStatus
	os         ovirtclient.VMOS
}

func (t *testVM) InstanceTypeID() *ovirtclient.InstanceTypeID {
	panic("not implemented for test input")
}

func (t *testVM) VMType() ovirtclient.VMType {
	panic("not implemented for test input")
}

func (t *testVM) OS() ovirtclient.VMOS {
	return t.os
}

func (t *testVM) Memory() int64 {
	panic("not implemented for test input")
}

func (t *testVM) MemoryPolicy() (ovirtclient.MemoryPolicy, bool) {
	panic("not implemented for test input")
}

func (t *testVM) TagIDs() []ovirtclient.TagID {
	panic("not implemented for test input")
}

func (t *testVM) HugePages() *ovirtclient.VMHugePages {
	panic("not implemented for test input")
}

func (t *testVM) Initialization() ovirtclient.Initialization {
	panic("not implemented for test input")
}

func (t *testVM) HostID() *ovirtclient.HostID {
	panic("not implemented for test input")
}

func (t *testVM) PlacementPolicy() (placementPolicy ovirtclient.VMPlacementPolicy, ok bool) {
	panic("not implemented for test input")
}

type testCPU struct {
	topo testTopo
}

func (t testCPU) Mode() *ovirtclient.CPUMode {
	panic("not implemented for test input")
}

type testTopo struct {
	cores   uint
	threads uint
	sockets uint
}

func (t testTopo) Cores() uint {
	return t.cores
}

func (t testTopo) Threads() uint {
	return t.threads
}

func (t testTopo) Sockets() uint {
	return t.sockets
}

func (t testCPU) Topo() ovirtclient.VMCPUTopo {
	return t.topo
}

func (t *testVM) CPU() ovirtclient.VMCPU {
	return testCPU{}
}

func (t *testVM) ID() ovirtclient.VMID {
	return t.id
}

func (t *testVM) Name() string {
	return t.name
}

func (t *testVM) Comment() string {
	return t.comment
}

func (t *testVM) ClusterID() ovirtclient.ClusterID {
	return t.clusterID
}

func (t *testVM) TemplateID() ovirtclient.TemplateID {
	return t.templateID
}

func (t *testVM) Status() ovirtclient.VMStatus {
	return t.status
}

type testOS struct {
	t string
}

func (t testOS) Type() string {
	return t.t
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
		os: &testOS{
			t: "linux",
		},
	}
	resourceData := schema.TestResourceDataRaw(t, vmSchema, map[string]interface{}{})
	diags := vmResourceUpdate(vm, resourceData)
	if len(diags) != 0 {
		t.Fatalf("failed to convert VM resource (%v)", diags)
	}
	compareResource(t, resourceData, "id", string(vm.id))
	compareResource(t, resourceData, "name", vm.name)
	compareResource(t, resourceData, "cluster_id", string(vm.clusterID))
	compareResource(t, resourceData, "template_id", string(vm.templateID))
	compareResource(t, resourceData, "status", string(vm.status))
	compareResource(t, resourceData, "os_type", vm.os.Type())
}

func compareResource(t *testing.T, data *schema.ResourceData, field string, value string) {
	if resourceValue := data.Get(field); resourceValue != value {
		t.Fatalf("invalid resource %s: %s, expected: %s", field, resourceValue, value)
	}
}
