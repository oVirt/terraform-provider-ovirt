package govirt

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

// TestHelper is a helper to run tests against an oVirt engine. When created it scans the oVirt Engine and tries to find
// working resources (hosts, clusters, etc) for running tests against. Tests should clean up after themselves.
type TestHelper interface {
	// GetClient returns the goVirt client.
	GetClient() Client

	// GetSDKClient returns the oVirt SDK client.
	GetSDKClient() *ovirtsdk4.Connection

	// GetClusterID returns the ID for the cluster.
	GetClusterID() string

	// GetBlankTemplateID returns the ID of the blank template that can be used for creating dummy VMs.
	GetBlankTemplateID() string

	// GetStorageDomainID returns the ID of the storage domain to create the images on.
	GetStorageDomainID() string

	// GenerateRandomID generates a random ID for testing.
	GenerateRandomID(length uint) string
}

func MustNewTestHelper(
	username string,
	password string,
	url string,
	insecure bool,
	caFile string,
	caBundle []byte,
	clusterID string,
	blankTemplateID string,
	storageDomainID string,
	logger Logger,
) TestHelper {
	helper, err := NewTestHelper(
		url,
		username,
		password,
		caFile,
		caBundle,
		insecure,
		clusterID,
		blankTemplateID,
		storageDomainID,
		logger,
	)
	if err != nil {
		panic(err)
	}
	return helper
}

func NewTestHelper(
	url string,
	username string,
	password string,
	caFile string,
	caBundle []byte,
	insecure bool,
	clusterID string,
	blankTemplateID string,
	storageDomainID string,
	logger Logger,
) (TestHelper, error) {
	client, err := New(
		url,
		username,
		password,
		caFile,
		caBundle,
		insecure,
		nil,
		logger,
	)
	if err != nil {
		return nil, err
	}

	if clusterID == "" {
		clusterID, err = findTestClusterID(client)
		if err != nil {
			return nil, fmt.Errorf("failed to find a cluster to test on (%w)", err)
		}
	} else {
		if err := verifyTestClusterID(client, clusterID); err != nil {
			return nil, fmt.Errorf("failed to verify cluster ID %s (%w)", clusterID, err)
		}
	}

	if storageDomainID == "" {
		storageDomainID, err = findTestStorageDomainID(client)
		if err != nil {
			return nil, fmt.Errorf("failed to find storage domain to test on (%w)", err)
		}
	} else {
		if err := verifyTestStorageDomainID(client, storageDomainID); err != nil {
			return nil, fmt.Errorf("failed to verify storage domain ID %s (%w)", storageDomainID, err)
		}
	}

	if blankTemplateID == "" {
		blankTemplateID, err = findBlankTemplateID(client)
		if err != nil {
			return nil, fmt.Errorf("failed to find blank template (%w)", err)
		}
	} else {
		if err := verifyBlankTemplateID(client, blankTemplateID); err != nil {
			return nil, fmt.Errorf("failed to verify blank template ID %s (%w)", blankTemplateID, err)
		}
	}

	return &testHelper{
		client:          client,
		clusterID:       clusterID,
		storageDomainID: storageDomainID,
		blankTemplateID: blankTemplateID,
		rand:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func findBlankTemplateID(client Client) (string, error) {
	templates, err := client.ListTemplates()
	if err != nil {
		return "", fmt.Errorf("failed to list templates (%w)", err)
	}
	for _, template := range templates {
		if strings.Contains(template.Description(), "Blank template") {
			return template.ID(), nil
		}
	}
	return "", fmt.Errorf("failed to find blank template for testing")
}

func verifyBlankTemplateID(client Client, templateID string) error {
	_, err := client.GetTemplate(templateID)
	return err
}

func findTestClusterID(client Client) (string, error) {
	clusters, err := client.ListClusters()
	if err != nil {
		return "", err
	}
	hosts, err := client.ListHosts()
	if err != nil {
		return "", err
	}
	for _, cluster := range clusters {
		for _, host := range hosts {
			if host.Status() == HostStatusUp && host.ClusterID() == cluster.ID() {
				return cluster.ID(), nil
			}
		}
	}
	return "", fmt.Errorf("failed to find cluster suitable for testing")
}

func verifyTestClusterID(client Client, clusterID string) error {
	_, err := client.GetCluster(clusterID)
	return err
}

func findTestStorageDomainID(client Client) (string, error) {
	storageDomains, err := client.ListStorageDomains()
	if err != nil {
		return "", err
	}
	for _, storageDomain := range storageDomains {
		// Assume 2GB will be enough for testing
		if storageDomain.Available() < 2*1024*1024*1024 {
			continue
		}
		if storageDomain.Status() != StorageDomainStatusActive &&
			storageDomain.ExternalStatus() != StorageDomainExternalStatusOk {
			continue
		}
		return storageDomain.ID(), nil
	}
	return "", fmt.Errorf("failed to find a working storage domain for testing")
}

func verifyTestStorageDomainID(client Client, storageDomainID string) error {
	_, err := client.GetStorageDomain(storageDomainID)
	return err
}

type testHelper struct {
	client          Client
	clusterID       string
	storageDomainID string
	rand            *rand.Rand
	blankTemplateID string
}

func (t *testHelper) GetClient() Client {
	return t.client
}

func (t *testHelper) GetSDKClient() *ovirtsdk4.Connection {
	return t.client.GetSDKClient()
}

func (t *testHelper) GetClusterID() string {
	return t.clusterID
}

func (t *testHelper) GetBlankTemplateID() string {
	return t.blankTemplateID
}

func (t *testHelper) GetStorageDomainID() string {
	return t.storageDomainID
}

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (t *testHelper) GenerateRandomID(length uint) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[t.rand.Intn(len(letters))]
	}
	return string(b)
}
