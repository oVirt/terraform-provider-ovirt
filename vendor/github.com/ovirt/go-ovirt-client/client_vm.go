package ovirtclient

import (
	"context"
	"fmt"
)

type VMClient interface {
	CreateVM(
		ctx context.Context,
		clusterID string,
		cpuTopo VMCPUTopo,
		templateID string,
		blockDevices []VMBlockDevice,
	)
}

// NewVMCPUTopo creates a new CPU topology with the given parameters. It returns an error if cores, threads, or sockets
// is 0. If the parameters are guaranteed to be non-zero MustNewVMCPUTopo should be used.
func NewVMCPUTopo(cores uint, threads uint, sockets uint) (VMCPUTopo, error) {
	if cores == 0 {
		return nil, fmt.Errorf("BUG: cores cannot be zero")
	}
	if threads == 0 {
		return nil, fmt.Errorf("BUG: threads cannot be zero")
	}
	if sockets == 0 {
		return nil, fmt.Errorf("BUG: sockets cannot be zero")
	}
	return &vmCPUTopo{
		cores:   cores,
		threads: threads,
		sockets: sockets,
	}, nil
}

// MustNewVMCPUTopo is identical to NewVMCPUTopo, but panics instead of returning an error if cores, threads, or
// sockets is zero.
func MustNewVMCPUTopo(cores uint, threads uint, sockets uint) VMCPUTopo {
	topo, err := NewVMCPUTopo(cores, threads, sockets)
	if err != nil {
		panic(err)
	}
	return topo
}

type VMCPUTopo interface {
	Cores() uint
	Threads() uint
	Sockets() uint
}

type vmCPUTopo struct {
	cores   uint
	threads uint
	sockets uint
}

func (v *vmCPUTopo) Cores() uint {
	return v.cores
}

func (v *vmCPUTopo) Threads() uint {
	return v.threads
}

func (v *vmCPUTopo) Sockets() uint {
	return v.sockets
}

type VMBlockDevice interface {
	DiskID() string
	Bootable() bool

	StorageDomainID() string
}
