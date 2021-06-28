package ovirtclient

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// MockClient provides in-memory client functions, and additionally provides the ability to inject
// information.
type MockClient interface {
	Client

	// GenerateUUID generates a UUID for testing purposes.
	GenerateUUID() string
}

type mockClient struct {
	url            string
	lock           *sync.Mutex
	storageDomains map[string]storageDomain
	disks          map[string]disk
	clusters       map[string]cluster
	hosts          map[string]host
	templates      map[string]template
}

func (m *mockClient) GetURL() string {
	return m.url
}

func (m *mockClient) GenerateUUID() string {
	return uuid.NewString()
}

func (m *mockClient) ListDisks() ([]Disk, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	result := make([]Disk, len(m.disks))
	i := 0
	for _, disk := range m.disks {
		result[i] = disk
		i++
	}
	return result, nil
}

func (m *mockClient) GetDisk(diskID string) (Disk, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if disk, ok := m.disks[diskID]; ok {
		return disk, nil
	}
	return nil, fmt.Errorf("disk with ID %s not found", diskID)
}

func (m *mockClient) RemoveDisk(diskID string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.disks[diskID]; ok {
		delete(m.disks, diskID)
		return nil
	}
	return fmt.Errorf("disk with ID %s not found", diskID)
}

func (m *mockClient) CreateVM(
	ctx context.Context,
	clusterID string,
	cpuTopo VMCPUTopo,
	templateID string,
	blockDevices []VMBlockDevice,
) {
	// TODO implement create VM
	panic("implement me")
}

func (m *mockClient) ListClusters() ([]Cluster, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	result := make([]Cluster, len(m.clusters))
	i := 0
	for _, c := range m.clusters {
		result[i] = c
		i++
	}
	return result, nil
}

func (m *mockClient) GetCluster(id string) (Cluster, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if c, ok := m.clusters[id]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("cluster with ID %s not found", id)
}

func (m *mockClient) ListStorageDomains() ([]StorageDomain, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	result := make([]StorageDomain, len(m.storageDomains))
	i := 0
	for _, c := range m.storageDomains {
		result[i] = c
		i++
	}
	return result, nil
}

func (m *mockClient) GetStorageDomain(id string) (StorageDomain, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if s, ok := m.storageDomains[id]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("storage domain with ID %s not found", id)
}

func (m *mockClient) ListHosts() ([]Host, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	result := make([]Host, len(m.hosts))
	i := 0
	for _, h := range m.hosts {
		result[i] = h
		i++
	}
	return result, nil
}

func (m *mockClient) GetHost(id string) (Host, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if h, ok := m.hosts[id]; ok {
		return h, nil
	}
	return nil, fmt.Errorf("host with ID %s not found", id)
}

func (m *mockClient) ListTemplates() ([]Template, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	result := make([]Template, len(m.templates))
	i := 0
	for _, t := range m.templates {
		result[i] = t
		i++
	}
	return result, nil
}

func (m *mockClient) GetTemplate(id string) (Template, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if t, ok := m.templates[id]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("template with ID %s not found", id)
}
