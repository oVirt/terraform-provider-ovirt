package ovirt

import (
    "context"
    "fmt"
    "regexp"
    "testing"

    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
    ovirtclient "github.com/ovirt/go-ovirt-client"
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
	cluster_id = "%s"
	template_id = "%s"
    name = "test"
    comment = "Hello world!"

	cpu_cores = 2
	cpu_threads = 3
	cpu_sockets = 4
    memory = 2147483648
    os_type = "rhcos_x64"
    initialization_custom_script = "echo 'Hello world!'"
    initialization_hostname = "test"
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
                            "comment",
                            regexp.MustCompile("^Hello world!$"),
                        ),
                        resource.TestMatchResourceAttr(
                            "ovirt_vm.foo",
                            "cpu_cores",
                            regexp.MustCompile("^2$"),
                        ),
                        resource.TestMatchResourceAttr(
                            "ovirt_vm.foo",
                            "cpu_threads",
                            regexp.MustCompile("^3$"),
                        ),
                        resource.TestMatchResourceAttr(
                            "ovirt_vm.foo",
                            "cpu_sockets",
                            regexp.MustCompile("^4$"),
                        ),
                        resource.TestMatchResourceAttr(
                            "ovirt_vm.foo",
                            "os_type",
                            regexp.MustCompile("^rhcos_x64$"),
                        ),
                        resource.TestMatchResourceAttr(
                            "ovirt_vm.foo",
                            "memory",
                            regexp.MustCompile("^2147483648$"),
                        ),
                        resource.TestMatchResourceAttr(
                            "ovirt_vm.foo",
                            "initialization_custom_script",
                            regexp.MustCompile("^echo 'Hello world!'$"),
                        ),
                        resource.TestMatchResourceAttr(
                            "ovirt_vm.foo",
                            "initialization_hostname",
                            regexp.MustCompile("^test$"),
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

    p := newProvider(newTestLogger(t))
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
                        vm, err := client.WithContext(context.Background()).CreateVM(
                            clusterID,
                            templateID,
                            "test",
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

type testVM struct {
    id             string
    name           string
    comment        string
    clusterID      string
    templateID     string
    status         ovirtclient.VMStatus
    os             testOSType
    memory         int64
    cpu            testCPU
    initialization testInitialization
}

type testInitialization struct {
    customScript string
    hostname     string
}

func (t testInitialization) CustomScript() string {
    return t.customScript
}

func (t testInitialization) HostName() string {
    return t.hostname
}

type testOSType struct {
    t string
}

func (t testOSType) Type() string {
    return t.t
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
    return t.memory
}

func (t *testVM) MemoryPolicy() (ovirtclient.MemoryPolicy, bool) {
    panic("not implemented for test input")
}

func (t *testVM) TagIDs() []string {
    panic("not implemented for test input")
}

func (t *testVM) HugePages() *ovirtclient.VMHugePages {
    panic("not implemented for test input")
}

func (t *testVM) Initialization() ovirtclient.Initialization {
    return t.initialization
}

func (t *testVM) HostID() *string {
    panic("not implemented for test input")
}

func (t *testVM) PlacementPolicy() (placementPolicy ovirtclient.VMPlacementPolicy, ok bool) {
    panic("not implemented for test input")
}

type testCPU struct {
    mode *ovirtclient.CPUMode
    topo testTopo
}

func (t testCPU) Mode() *ovirtclient.CPUMode {
    return t.mode
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
    return t.cpu
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

func (t *testVM) ClusterID() ovirtclient.ClusterID {
    return ovirtclient.ClusterID(t.clusterID)
}

func (t *testVM) TemplateID() ovirtclient.TemplateID {
    return ovirtclient.TemplateID(t.templateID)
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
        os: testOSType{
            t: "rhcos_x64",
        },
        memory: 1024 * 1024 * 1024,
        cpu: testCPU{
            topo: testTopo{
                cores:   1,
                threads: 1,
                sockets: 1,
            },
        },
    }
    resourceData := schema.TestResourceDataRaw(t, vmSchema, map[string]interface{}{})
    diags := vmResourceUpdate(vm, resourceData)
    if len(diags) != 0 {
        t.Fatalf("failed to convert VM resource (%v)", diags)
    }
    compareResource(t, resourceData, "id", vm.id)
    compareResource(t, resourceData, "name", vm.name)
    compareResource(t, resourceData, "comment", vm.comment)
    compareResource(t, resourceData, "cluster_id", vm.clusterID)
    compareResource(t, resourceData, "template_id", vm.templateID)
    compareResource(t, resourceData, "status", string(vm.status))
    compareResourceStringPointer(t, resourceData, "cpu_mode", (*string)(vm.cpu.mode))
    compareResourceInt(t, resourceData, "memory", vm.memory)
    compareResourceUInt(t, resourceData, "cpu_threads", vm.cpu.topo.cores)
    compareResourceUInt(t, resourceData, "cpu_sockets", vm.cpu.topo.cores)
    compareResourceUInt(t, resourceData, "cpu_cores", vm.cpu.topo.cores)

    rawOS, ok := resourceData.GetOk("os")
    if ok {
        os := rawOS.(*schema.Set)
        if len(os.List()) == 0 {
            t.Fatalf("OS type not set")
        }
        if val := os.List()[0].(map[string]interface{})["type"]; val != "rhcos_x64" {
            t.Fatalf("invalid resource %s: %s, expected: rhcos_64", "os[0].type", val)
        }
    }
}

func compareResourceUInt(t *testing.T, data *schema.ResourceData, field string, value uint) {
    resourceValue, ok := data.GetOk(field)
    if !ok {
        t.Fatalf("field %s not set", field)
    }
    if uint(resourceValue.(int)) != value {
        t.Fatalf("invalid resource %s: %s, expected: %d", field, resourceValue, value)
    }
}

func compareResourceInt(t *testing.T, data *schema.ResourceData, field string, value int64) {
    resourceValue, ok := data.GetOk(field)
    if !ok {
        t.Fatalf("field %s not set", field)
    }
    actual := int64(resourceValue.(int))
    if actual != value {
        t.Fatalf("invalid resource %s: %d, expected: %d", field, actual, value)
    }
}

func compareResource(t *testing.T, data *schema.ResourceData, field string, value string) {
    resourceValue, ok := data.GetOk(field)
    if !ok {
        t.Fatalf("field %s not set", field)
    }
    if resourceValue != value {
        t.Fatalf("invalid resource %s: %s, expected: %s", field, resourceValue, value)
    }
}

func compareResourceStringPointer(t *testing.T, data *schema.ResourceData, field string, value *string) {
    resourceValue, ok := data.GetOk(field)
    if !ok {
        if value != nil {
            t.Fatalf("field %s not set", field)
        }
        return
    }
    if value == nil {
        t.Fatalf("invalid resource %s: %s, expected: nil", field, resourceValue)
    }
    if resourceValue != *value {
        t.Fatalf("invalid resource %s: %s, expected: %s", field, resourceValue, *value)
    }
}
