//nolint:revive
package ovirt

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v3"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
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

func TestVMResourceVMType(t *testing.T) {
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
    vm_type     = "server"
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
						func(state *terraform.State) error {
							VMID := state.RootModule().Resources["ovirt_vm.foo"].Primary.ID
							vm, err := p.getTestHelper().GetClient().GetVM(ovirtclient.VMID(VMID))
							if err != nil {
								return fmt.Errorf("Failed to get VM: %w", err)
							}
							if vm.VMType() != ovirtclient.VMTypeServer {
								return fmt.Errorf("Expected VM type to be %s, but got %s", ovirtclient.VMTypeServer, vm.VMType())
							}
							return nil
						},
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

func TestVMResourceInitialization(t *testing.T) {
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
	initialization_hostname = "vm-test-1"
	initialization_custom_script = "echo hello"
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
							"initialization_hostname",
							regexp.MustCompile("^vm-test-1$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_vm.foo",
							"initialization_custom_script",
							regexp.MustCompile("^echo hello$"),
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

func TestVMResourceInitializationWithNicConfiguration(t *testing.T) {
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
	initialization_hostname = "vm-test-1"
	initialization_custom_script = "echo hello"
	initialization_nic {
		name = "eth0"
		ipv4 {
			address = "1.2.3.4"
			netmask = "255.255.255.0"
			gateway = "1.2.3.1"
		}
	}
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
							"initialization_hostname",
							regexp.MustCompile("^vm-test-1$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_vm.foo",
							"initialization_custom_script",
							regexp.MustCompile("^echo hello$"),
						),
					),
				},
				{
					Config: config,
					Check: func(state *terraform.State) error {
						vmID := state.RootModule().Resources["ovirt_vm.foo"].Primary.ID
						vm, err := p.getTestHelper().GetClient().GetVM(ovirtclient.VMID(vmID))
						if err != nil {
							return err
						}
						initialization := vm.Initialization()
						if initialization == nil {
							return fmt.Errorf("Initialization not set")
						}
						nicConfiguration := initialization.NicConfiguration()
						if nicConfiguration == nil {
							return fmt.Errorf("No NIC configuration in initialization")
						}
						if nicConfiguration.Name() != "eth0" {
							return fmt.Errorf("incorrect NIC name: %s", nicConfiguration.Name())
						}

						ipConfiguration := nicConfiguration.IP()
						if ipConfiguration.Address != "1.2.3.4" {
							return fmt.Errorf("incorrect IP address: %s", ipConfiguration.Address)
						}
						if ipConfiguration.Netmask != "255.255.255.0" {
							return fmt.Errorf("incorrect IP address: %s", ipConfiguration.Netmask)
						}
						if ipConfiguration.Gateway != "1.2.3.1" {
							return fmt.Errorf("incorrect IP address: %s", ipConfiguration.Gateway)
						}
						if nicConfiguration.IPV6() != nil {
							return fmt.Errorf("Found IPV6 configuration in initialization_nic")
						}
						return nil
					},
				},
				{
					Config:  config,
					Destroy: true,
				},
			},
		},
	)
}
func TestVMResourceInitializationWithNicConfigurationV6(t *testing.T) {
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
	initialization_hostname = "vm-test-1"
	initialization_custom_script = "echo hello"
	initialization_nic {
		name = "eth0"
		ipv4 {
			address = "1.2.3.4"
			netmask = "255.255.255.0"
			gateway = "1.2.3.1"
		}
		ipv6 {
			address = "2001:DB8::12"
			netmask = "64"
			gateway = "2001:DB8::1"
		}
	}
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
							"initialization_hostname",
							regexp.MustCompile("^vm-test-1$"),
						),
						resource.TestMatchResourceAttr(
							"ovirt_vm.foo",
							"initialization_custom_script",
							regexp.MustCompile("^echo hello$"),
						),
					),
				},
				{
					Config: config,
					Check: func(state *terraform.State) error {
						vmID := state.RootModule().Resources["ovirt_vm.foo"].Primary.ID
						vm, err := p.getTestHelper().GetClient().GetVM(ovirtclient.VMID(vmID))
						if err != nil {
							return err
						}
						initialization := vm.Initialization()
						if initialization == nil {
							return fmt.Errorf("Initialization not set")
						}
						nicConfiguration := initialization.NicConfiguration()
						if nicConfiguration == nil {
							return fmt.Errorf("No NIC configuration in initialization")
						}
						if nicConfiguration.Name() != "eth0" {
							return fmt.Errorf("incorrect NIC name: %s", nicConfiguration.Name())
						}

						ipConfiguration := nicConfiguration.IP()
						if ipConfiguration.Address != "1.2.3.4" {
							return fmt.Errorf("incorrect IP address: %s", ipConfiguration.Address)
						}
						if ipConfiguration.Netmask != "255.255.255.0" {
							return fmt.Errorf("incorrect IP address: %s", ipConfiguration.Netmask)
						}
						if ipConfiguration.Gateway != "1.2.3.1" {
							return fmt.Errorf("incorrect IP address: %s", ipConfiguration.Gateway)
						}
						ipv6Configuration := nicConfiguration.IPV6()
						if ipv6Configuration == nil {
							return fmt.Errorf("Could not find IPV6 configuration in initialization_nic")
						}
						if ipv6Configuration.Address != "2001:DB8::12" {
							return fmt.Errorf("incorrect IP address: %s", ipv6Configuration.Address)
						}
						if ipv6Configuration.Netmask != "64" {
							return fmt.Errorf("incorrect IP address: %s", ipv6Configuration.Netmask)
						}
						if ipv6Configuration.Gateway != "2001:DB8::1" {
							return fmt.Errorf("incorrect IP address: %s", ipv6Configuration.Gateway)
						}
						return nil
					},
				},
				{
					Config:  config,
					Destroy: true,
				},
			},
		},
	)
}

func TestVMResourceCPUParameters(t *testing.T) {
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
	cpu_cores   = 2
	cpu_threads = 2
	cpu_sockets = 2
	cpu_mode    = "host_passthrough"
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
					Check: func(state *terraform.State) error {
						vmID := state.RootModule().Resources["ovirt_vm.foo"].Primary.ID
						vm, err := p.getTestHelper().GetClient().GetVM(ovirtclient.VMID(vmID))
						if err != nil {
							return err
						}
						mode := vm.CPU().Mode()
						if mode == nil {
							return fmt.Errorf("CPU mode not set")
						}
						if *mode != ovirtclient.CPUModeHostPassthrough {
							return fmt.Errorf("incorrect CPU mode: %s", *mode)
						}
						if cores := vm.CPU().Topo().Cores(); cores != 2 {
							return fmt.Errorf("incorrect number of CPU cores: %d", cores)
						}
						if threads := vm.CPU().Topo().Threads(); threads != 2 {
							return fmt.Errorf("incorrect number of CPU threads: %d", threads)
						}
						if sockets := vm.CPU().Topo().Sockets(); sockets != 2 {
							return fmt.Errorf("incorrect number of CPU sockets: %d", sockets)
						}
						return nil
					},
				},
				{
					Config:  config,
					Destroy: true,
				},
			},
		},
	)
}

func TestVMResourcePlacementPolicy(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	clusterID := p.getTestHelper().GetClusterID()
	templateID := p.getTestHelper().GetBlankTemplateID()

	client := p.getTestHelper().GetClient().WithContext(context.Background())
	hosts, err := client.ListHosts()
	if err != nil {
		t.Fatalf("Failed to list hosts (%v)", err)
	}
	hostIDs := make([]string, len(hosts))
	for i, host := range hosts {
		hostIDs[i] = fmt.Sprintf("\"%s\"", host.ID())
	}
	config := fmt.Sprintf(
		`
provider "ovirt" {
	mock = true
}

resource "ovirt_vm" "foo" {
	cluster_id                = "%s"
	template_id               = "%s"
	name                      = "test"
	placement_policy_affinity = "migratable"
	placement_policy_host_ids = [%s]
}
`,
		clusterID,
		templateID,
		strings.Join(hostIDs, ","),
	)

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: func(state *terraform.State) error {
						vmData := state.RootModule().Resources["ovirt_vm.foo"]
						vm, err := client.GetVM(ovirtclient.VMID(vmData.Primary.ID))
						if err != nil {
							return fmt.Errorf("failed to fetch VM (%w)", err)
						}
						placementPolicy, ok := vm.PlacementPolicy()
						if !ok {
							return fmt.Errorf("no placement policy on the VM")
						}
						if *placementPolicy.Affinity() != ovirtclient.VMAffinityMigratable {
							return fmt.Errorf("incorrect VM affinity: %s", *placementPolicy.Affinity())
						}
						return nil
					},
				},
				{
					Config:  config,
					Destroy: true,
				},
			},
		},
	)
}

func TestVMResourceInstanceTypeID(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	clusterID := p.getTestHelper().GetClusterID()
	templateID := p.getTestHelper().GetBlankTemplateID()
	client := p.getTestHelper().GetClient().WithContext(context.Background())

	instanceTypes, err := client.ListInstanceTypes()
	testHelper := p.getTestHelper()

	if err != nil {
		t.Fatalf("Failed to list instance types (%v)", err)
	}
	instanceType := instanceTypes[0].ID()

	config := fmt.Sprintf(
		`
provider "ovirt" {
	mock = true
}

resource "ovirt_vm" "foo" {
	cluster_id  = "%s"
	template_id = "%s"
	name        = "test"
	instance_type_id     = "%s"
}
`,
		clusterID,
		templateID,
		instanceType,
	)

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: func(state *terraform.State) error {
						client := testHelper.GetClient()
						vmID := state.RootModule().Resources["ovirt_vm.foo"].Primary.ID
						vm, err := client.GetVM(ovirtclient.VMID(vmID))
						if err != nil {
							return err
						}
						if *vm.InstanceTypeID() != instanceType {
							return fmt.Errorf("incorrect value for instance Type ID: %s ", *vm.InstanceTypeID())
						}
						return nil
					},
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
	id              ovirtclient.VMID
	name            string
	comment         string
	clusterID       ovirtclient.ClusterID
	templateID      ovirtclient.TemplateID
	status          ovirtclient.VMStatus
	os              ovirtclient.VMOS
	placementPolicy ovirtclient.VMPlacementPolicy
}

func (t *testVM) Description() string {
	panic("not implemented for test input")
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

func (t *testVM) MemoryPolicy() ovirtclient.MemoryPolicy {
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
	return t.placementPolicy, t.placementPolicy != nil
}

type testPlacementPolicy struct {
	affinity *ovirtclient.VMAffinity
	hostIDs  []ovirtclient.HostID
}

func (t testPlacementPolicy) Affinity() *ovirtclient.VMAffinity {
	return t.affinity
}

func (t testPlacementPolicy) HostIDs() []ovirtclient.HostID {
	return t.hostIDs
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

	vmAffinity := ovirtclient.VMAffinityMigratable
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
		placementPolicy: &testPlacementPolicy{
			&vmAffinity,
			[]ovirtclient.HostID{"asdf"},
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
	compareResource(t, resourceData, "placement_policy_affinity", string(*vm.placementPolicy.Affinity()))
	compareResourceStringList(t, resourceData, "placement_policy_host_ids", []string{"asdf"})
}

func compareResource(t *testing.T, data *schema.ResourceData, field string, value string) {
	if resourceValue := data.Get(field); resourceValue != value {
		t.Fatalf("invalid resource %s: %s, expected: %s", field, resourceValue, value)
	}
}

func compareResourceStringList(t *testing.T, data *schema.ResourceData, field string, expectedValues []string) {
	resourceValue := data.Get(field).(*schema.Set)
	realValues := resourceValue.List()
	if len(realValues) != len(expectedValues) {
		t.Fatalf("Incorrect number of values found (expected: %d, found: %d)", len(expectedValues), len(realValues))
	}
	for _, value := range realValues {
		found := false
		for _, expectedValue := range expectedValues {
			if expectedValue == value {
				found = true
			}
		}
		if !found {
			t.Fatalf("Invalid value found: %s", value)
		}
	}
}

func TestVMOverrideDisk(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	testHelper := p.getTestHelper()
	clusterID := testHelper.GetClusterID()
	templateID := testHelper.GetBlankTemplateID()
	storageDomainID := testHelper.GetStorageDomainID()
	secondaryStorageDomainID := testHelper.GetSecondaryStorageDomainID(t)

	baseConfig := fmt.Sprintf(
		`
		provider "ovirt" {
			mock = true
		}

		resource "ovirt_vm" "source" {
			cluster_id  = "%s"
			template_id = "%s"
			name        = "%s"
		}

		resource "ovirt_disk" "source" {
			storage_domain_id = "%s"
			format           = "cow"
			size             = 1048576
			alias            = "test"
			sparse           = false
		}

		resource "ovirt_disk_attachment" "source" {
			vm_id          = ovirt_vm.source.id
			disk_id        = ovirt_disk.source.id
			disk_interface = "virtio_scsi"
		}

		resource "ovirt_template" "source" {
			vm_id = ovirt_disk_attachment.source.vm_id
			name  = "%s"
		}

		data "ovirt_template_disk_attachments" "source" {
			template_id = ovirt_template.source.id
		}`,
		clusterID,
		templateID,
		p.getTestHelper().GenerateTestResourceName(t),
		storageDomainID,
		p.getTestHelper().GenerateTestResourceName(t),
	)

	rconf :=
		`
		resource "ovirt_vm" "one" {
			template_id = ovirt_template.source.id
			cluster_id  = "%s"
			name        = "%s"
			clone       = true

			dynamic "template_disk_attachment_override" {
				for_each = data.ovirt_template_disk_attachments.source.disk_attachments
				content {
					disk_id           = template_disk_attachment_override.value["disk_id"]
					format            = %s
					provisioning      = %s
					storage_domain_id = %s
				}
			}
		}`

	testcases := []struct {
		name                    string
		inputFormat             string
		inputProvisioning       string
		inputStorageDomainID    string
		expectedFormat          ovirtclient.ImageFormat
		expectedSparse          bool
		expectedStorageDomainID string
	}{
		{
			name:                    "override sparse",
			inputFormat:             "null",
			inputProvisioning:       "\"sparse\"",
			inputStorageDomainID:    "null",
			expectedFormat:          ovirtclient.ImageFormatCow,
			expectedSparse:          true,
			expectedStorageDomainID: string(storageDomainID),
		},
		{
			name:                    "override format",
			inputFormat:             "\"raw\"",
			inputProvisioning:       "null",
			inputStorageDomainID:    "null",
			expectedFormat:          ovirtclient.ImageFormatRaw,
			expectedSparse:          false,
			expectedStorageDomainID: string(storageDomainID),
		},
		{
			name:                    "set storage_domain_id",
			inputFormat:             "null",
			inputProvisioning:       "null",
			inputStorageDomainID:    "\"" + string(secondaryStorageDomainID) + "\"",
			expectedFormat:          ovirtclient.ImageFormatCow,
			expectedSparse:          false,
			expectedStorageDomainID: string(secondaryStorageDomainID),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			config := baseConfig + fmt.Sprintf(rconf,
				clusterID,
				p.getTestHelper().GenerateTestResourceName(t),
				testcase.inputFormat,
				testcase.inputProvisioning,
				testcase.inputStorageDomainID,
			)

			resource.UnitTest(
				t, resource.TestCase{
					ProviderFactories: p.getProviderFactories(),
					Steps: []resource.TestStep{
						{
							Config: config,
							Check: func(state *terraform.State) error {
								client := testHelper.GetClient()
								vmID := state.RootModule().Resources["ovirt_vm.one"].Primary.ID
								diskAttachments, err := client.ListDiskAttachments(ovirtclient.VMID(vmID))
								if err != nil {
									return err
								}
								diskAttachment := diskAttachments[0]
								disk, err := diskAttachment.Disk()
								if err != nil {
									return err
								}
								if disk.Format() != testcase.expectedFormat {
									return fmt.Errorf("incorrect disk format: %s", disk.Format())
								}
								if disk.Sparse() != testcase.expectedSparse {
									return fmt.Errorf("disk incorrectly created as sparse")
								}
								storageDomainID := disk.StorageDomainIDs()[0]
								if string(storageDomainID) != testcase.expectedStorageDomainID {
									return fmt.Errorf("disk not created on correct storage domain")
								}
								return nil
							},
						},
						{
							Config:  config,
							Destroy: true,
						},
					},
				},
			)
		})
	}
}

func TestMemory(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	testHelper := p.getTestHelper()
	clusterID := testHelper.GetClusterID()
	config := fmt.Sprintf(
		`
provider "ovirt" {
	mock = true
}

data "ovirt_blank_template" "blank" {
}

resource "ovirt_vm" "test" {
	template_id       = data.ovirt_blank_template.blank.id
	cluster_id        = "%s"
	name              = "%s"
	memory            = 1048576
	maximum_memory    = 2097152
    memory_ballooning = false
}
`,
		clusterID,
		p.getTestHelper().GenerateTestResourceName(t),
	)

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: func(state *terraform.State) error {
						client := testHelper.GetClient()
						vmID := state.RootModule().Resources["ovirt_vm.test"].Primary.ID
						vm, err := client.GetVM(ovirtclient.VMID(vmID))
						if err != nil {
							return err
						}
						if vm.Memory() != 1048576 {
							return fmt.Errorf("incorrect amount of memory: %d", vm.Memory())
						}
						memoryPolicy := vm.MemoryPolicy()
						if memoryPolicy.Max() == nil {
							return fmt.Errorf("no maximum memory set on VM")
						}
						if *memoryPolicy.Max() != 2097152 {
							return fmt.Errorf("incorrect maximum memory set on VM: %d", *memoryPolicy.Max())
						}

						return nil
					},
				},
				{
					Config:  config,
					Destroy: true,
				},
			},
		},
	)
}

func TestMemoryBallooning(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	testHelper := p.getTestHelper()
	clusterID := testHelper.GetClusterID()

	baseConfig := `
		provider "ovirt" {
			mock = true
		}

		data "ovirt_blank_template" "blank" {
		}

		resource "ovirt_vm" "test" {
			template_id       = data.ovirt_blank_template.blank.id
			cluster_id        = "%s"
			name              = "%s"
			memory_ballooning = %s
		}`

	testcases := []struct {
		inputMemoryBalloning     string
		expectedMemoryBallooning bool
	}{
		{
			inputMemoryBalloning:     "null",
			expectedMemoryBallooning: true,
		},
		{
			inputMemoryBalloning:     "false",
			expectedMemoryBallooning: false,
		},
		{
			inputMemoryBalloning:     "true",
			expectedMemoryBallooning: true,
		},
	}

	for _, tc := range testcases {
		tcName := fmt.Sprintf("memory_ballooning_for_%s", tc.inputMemoryBalloning)
		t.Run(tcName, func(t *testing.T) {
			config := fmt.Sprintf(
				baseConfig,
				clusterID,
				p.getTestHelper().GenerateTestResourceName(t),
				tc.inputMemoryBalloning,
			)

			resource.UnitTest(
				t, resource.TestCase{
					ProviderFactories: p.getProviderFactories(),
					Steps: []resource.TestStep{
						{
							Config: config,
							Check: func(state *terraform.State) error {
								client := testHelper.GetClient()
								vmID := state.RootModule().Resources["ovirt_vm.test"].Primary.ID
								vm, err := client.GetVM(ovirtclient.VMID(vmID))
								if err != nil {
									return err
								}

								if vm.MemoryPolicy().Ballooning() != tc.expectedMemoryBallooning {
									return fmt.Errorf("Expected memory_ballooning to be %t, but got %t",
										tc.expectedMemoryBallooning,
										vm.MemoryPolicy().Ballooning())
								}

								return nil
							},
						},
						{
							Config:  config,
							Destroy: true,
						},
					},
				},
			)
		})
	}
}

func TestSoundcardEnabled(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	testHelper := p.getTestHelper()
	clusterID := testHelper.GetClusterID()

	baseConfig := `
		provider "ovirt" {
			mock = true
		}

		data "ovirt_blank_template" "blank" {
		}

		resource "ovirt_vm" "test" {
			template_id       = data.ovirt_blank_template.blank.id
			cluster_id        = "%s"
			name              = "%s"
			soundcard_enabled = %s
		}`

	testcases := []struct {
		inputSoundcardEnabled    string
		expectedSoundcardEnabled bool
	}{
		{
			inputSoundcardEnabled:    "null",
			expectedSoundcardEnabled: true,
		},
		{
			inputSoundcardEnabled:    "false",
			expectedSoundcardEnabled: false,
		},
		{
			inputSoundcardEnabled:    "true",
			expectedSoundcardEnabled: true,
		},
	}

	for _, tc := range testcases {
		tcName := fmt.Sprintf("soundcard_enabled_for_%s", tc.inputSoundcardEnabled)
		t.Run(tcName, func(t *testing.T) {
			config := fmt.Sprintf(
				baseConfig,
				clusterID,
				p.getTestHelper().GenerateTestResourceName(t),
				tc.inputSoundcardEnabled,
			)

			resource.UnitTest(
				t, resource.TestCase{
					ProviderFactories: p.getProviderFactories(),
					Steps: []resource.TestStep{
						{
							Config: config,
							Check: func(state *terraform.State) error {
								client := testHelper.GetClient()
								vmID := state.RootModule().Resources["ovirt_vm.test"].Primary.ID
								vm, err := client.GetVM(ovirtclient.VMID(vmID))
								if err != nil {
									return err
								}

								if vm.SoundcardEnabled() != tc.expectedSoundcardEnabled {
									return fmt.Errorf("Expected soundcard_enabled to be %t, but got %t",
										tc.expectedSoundcardEnabled,
										vm.SoundcardEnabled())
								}

								return nil
							},
						},
						{
							Config:  config,
							Destroy: true,
						},
					},
				},
			)
		})
	}
}

func TestSerialConsole(t *testing.T) {
	t.Parallel()
	no := false
	yes := true
	testCases := []struct {
		set      *bool
		expected bool
	}{
		{
			nil,
			false,
		},
		{
			&no,
			false,
		},
		{
			&yes,
			true,
		},
	}

	for _, tc := range testCases {
		set := "nil"
		if tc.set != nil {
			set = fmt.Sprintf("%t", *tc.set)
		}
		expected := fmt.Sprintf("%t", tc.expected)
		t.Run(
			fmt.Sprintf("serial_console=%s,expected=%s", set, expected), func(t *testing.T) {
				p := newProvider(newTestLogger(t))
				testHelper := p.getTestHelper()
				clusterID := testHelper.GetClusterID()
				serialConsoleLine := ""
				if tc.set != nil {
					serialConsoleLine = fmt.Sprintf("serial_console = %t", *tc.set)
				}
				config := fmt.Sprintf(
					`
provider "ovirt" {
	mock = true
}

data "ovirt_blank_template" "blank" {
}

resource "ovirt_vm" "test" {
	template_id    = data.ovirt_blank_template.blank.id
	cluster_id     = "%s"
	name           = "%s"
    %s
}
`,
					clusterID,
					p.getTestHelper().GenerateTestResourceName(t),
					serialConsoleLine,
				)

				resource.UnitTest(
					t, resource.TestCase{
						ProviderFactories: p.getProviderFactories(),
						Steps: []resource.TestStep{
							{
								Config: config,
								Check: func(state *terraform.State) error {
									client := testHelper.GetClient()
									vmID := state.RootModule().Resources["ovirt_vm.test"].Primary.ID
									vm, err := client.GetVM(ovirtclient.VMID(vmID))
									if err != nil {
										return err
									}
									if vm.SerialConsole() != tc.expected {
										return fmt.Errorf("incorrect value for serial console: %t", vm.SerialConsole())
									}

									return nil
								},
							},
							{
								Config:  config,
								Destroy: true,
							},
						},
					},
				)
			},
		)
	}
}

func TestClone(t *testing.T) {
	t.Parallel()
	no := false
	yes := true
	testCases := []struct {
		set                   *bool
		expectBlankTemplateID bool
	}{
		{
			nil,
			false,
		},
		{
			&no,
			false,
		},
		{
			&yes,
			true,
		},
	}

	for _, tc := range testCases {
		set := "nil"
		if tc.set != nil {
			set = fmt.Sprintf("%t", *tc.set)
		}
		expectBlank := fmt.Sprintf("%t", tc.expectBlankTemplateID)
		t.Run(
			fmt.Sprintf("clone=%s,expect blank template ID=%s", set, expectBlank), func(t *testing.T) {
				p := newProvider(newTestLogger(t))
				testHelper := p.getTestHelper()
				clusterID := testHelper.GetClusterID()
				cloneLine := ""
				if tc.set != nil {
					cloneLine = fmt.Sprintf("clone = %t", *tc.set)
				}
				config := fmt.Sprintf(
					`
provider "ovirt" {
	mock = true
}
data "ovirt_blank_template" "blank" {
}
resource "ovirt_vm" "test" {
	template_id    = data.ovirt_blank_template.blank.id
	cluster_id     = "%s"
	name           = "%s"
}
resource "ovirt_template" "blueprint" {
	vm_id	= ovirt_vm.test.id
	name	= "blueprint1"
}
resource "ovirt_vm" "second_vm" {
	template_id    = ovirt_template.blueprint.id
	cluster_id     = "%s"
	name           = "%s"
    %s
}
`,
					clusterID,
					p.getTestHelper().GenerateTestResourceName(t),
					clusterID,
					p.getTestHelper().GenerateTestResourceName(t),
					cloneLine,
				)

				resource.UnitTest(
					t, resource.TestCase{
						ProviderFactories: p.getProviderFactories(),
						Steps: []resource.TestStep{
							{
								Config: config,
								Check: func(state *terraform.State) error {
									secondVMID := state.RootModule().Resources["ovirt_vm.second_vm"].Primary.ID
									templateID := state.RootModule().Resources["ovirt_template.blueprint"].Primary.ID

									client := testHelper.GetClient()
									secondVM, err := client.GetVM(ovirtclient.VMID(secondVMID))
									if err != nil {
										return err
									}
									secondVMTemplateID := string(secondVM.TemplateID())
									if tc.expectBlankTemplateID && (secondVMTemplateID != string(ovirtclient.DefaultBlankTemplateID)) {
										return fmt.Errorf(
											"Expected template of VM to be blank '%s', but got '%s",
											string(ovirtclient.DefaultBlankTemplateID), secondVMTemplateID)
									}
									if !tc.expectBlankTemplateID && (secondVMTemplateID != templateID) {
										return fmt.Errorf("Expected template of VM to be '%s', but got '%s", templateID, secondVMTemplateID)
									}

									return nil
								},
							},
							{
								Config:  config,
								Destroy: true,
							},
						},
					},
				)
			},
		)
	}
}

func TestHugePages(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	testHelper := p.getTestHelper()
	clusterID := testHelper.GetClusterID()

	config := fmt.Sprintf(`
		provider "ovirt" {
			mock = true
		}
		data "ovirt_blank_template" "blank" {
		}
		resource "ovirt_vm" "test" {
			template_id    = data.ovirt_blank_template.blank.id
			cluster_id     = "%s"
			name           = "%s"
			huge_pages     = %d
		}`,
		clusterID,
		p.getTestHelper().GenerateTestResourceName(t),
		1048576)

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: func(state *terraform.State) error {
						VMID := state.RootModule().Resources["ovirt_vm.test"].Primary.ID

						client := testHelper.GetClient()
						VM, err := client.GetVM(ovirtclient.VMID(VMID))
						if err != nil {
							return err
						}
						if VM.HugePages() == nil {
							return fmt.Errorf("Expected huge pages to be set, but got <nil>")
						}
						if *VM.HugePages() != ovirtclient.VMHugePages1G {
							return fmt.Errorf("Expected huge pages to be %d, but got %d",
								ovirtclient.VMHugePages1G, *VM.HugePages())
						}

						return nil
					},
				},
				{
					Config:  config,
					Destroy: true,
				},
			},
		},
	)
}
