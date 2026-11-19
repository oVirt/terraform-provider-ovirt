package ovirt

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

// testVMCPUTopo mirrors the unexported topology type of go-ovirt-client: the accessors have a pointer
// receiver and dereference it, so calling them on a nil pointer panics.
type testVMCPUTopo struct {
	cores   uint
	threads uint
	sockets uint
}

func (t *testVMCPUTopo) Cores() uint   { return t.cores }
func (t *testVMCPUTopo) Threads() uint { return t.threads }
func (t *testVMCPUTopo) Sockets() uint { return t.sockets }

// testVMCPU returns its topology as an interface value. A nil topo therefore yields a non-nil interface
// wrapping a nil pointer, exactly as go-ovirt-client does for a VM that carries no topology.
type testVMCPU struct {
	topo *testVMCPUTopo
}

func (c testVMCPU) Topo() ovirtclient.VMCPUTopo { return c.topo }
func (c testVMCPU) Mode() *ovirtclient.CPUMode  { return nil }

func TestVMCPUResourceUpdateNilTopology(t *testing.T) {
	t.Parallel()

	data := schema.TestResourceDataRaw(t, vmSchema, map[string]interface{}{
		"cpu_cores":   4,
		"cpu_threads": 1,
		"cpu_sockets": 2,
	})

	// A nil topology must be skipped rather than dereferenced.
	diags := vmCPUResourceUpdate(testVMCPU{}, data, diag.Diagnostics{})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
}

func TestVMCPUResourceUpdateWritesTopology(t *testing.T) {
	t.Parallel()

	data := schema.TestResourceDataRaw(t, vmSchema, map[string]interface{}{
		"cpu_cores":   1,
		"cpu_threads": 1,
		"cpu_sockets": 1,
	})

	cpu := testVMCPU{topo: &testVMCPUTopo{cores: 4, threads: 2, sockets: 8}}
	diags := vmCPUResourceUpdate(cpu, data, diag.Diagnostics{})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	for field, want := range map[string]int{"cpu_cores": 4, "cpu_threads": 2, "cpu_sockets": 8} {
		if got := data.Get(field).(int); got != want {
			t.Errorf("%s = %d, want %d", field, got, want)
		}
	}
}
