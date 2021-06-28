package ovirtclient

import (
	"sync"

	"github.com/google/uuid"
)

// NewMock creates a new in-memory mock client. This client can be used as a testing facility for
// higher level code.
func NewMock() MockClient {
	testCluster := generateTestCluster()
	testHost := generateTestHost(testCluster)
	testStorageDomain := generateTestStorageDomain()

	return &mockClient{
		url:  "https://localhost/ovirt-engine/api",
		lock: &sync.Mutex{},
		storageDomains: map[string]storageDomain{
			testStorageDomain.ID(): testStorageDomain,
		},
		disks: map[string]disk{},
		clusters: map[string]cluster{
			testCluster.ID(): testCluster,
		},
		hosts: map[string]host{
			testHost.ID(): testHost,
		},
		templates: map[string]template{
			BlankTemplateID: {
				id:          BlankTemplateID,
				name:        "Blank",
				description: "Blank template",
			},
		},
	}
}

func generateTestStorageDomain() storageDomain {
	return storageDomain{
		id:             uuid.NewString(),
		name:           "Test storage domain",
		available:      10 * 1024 * 1024 * 1024,
		status:         StorageDomainStatusActive,
		externalStatus: StorageDomainExternalStatusNA,
	}
}

func generateTestCluster() cluster {
	return cluster{
		id:   uuid.NewString(),
		name: "Test cluster",
	}
}

func generateTestHost(c cluster) host {
	return host{
		id:        uuid.NewString(),
		clusterID: c.ID(),
		status:    HostStatusUp,
	}
}
