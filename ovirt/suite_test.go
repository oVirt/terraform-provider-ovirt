package ovirt_test

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	govirt "github.com/ovirt/go-ovirt-client"
	ovirtsdk4 "github.com/ovirt/go-ovirt"

	"github.com/ovirt/terraform-provider-ovirt/ovirt"
)

func getOvirtTestSuite(t *testing.T) OvirtTestSuite {
	suite, err := NewOvirtTestSuite(
		t,
		os.Getenv("OVIRT_USERNAME"),
		os.Getenv("OVIRT_PASSWORD"),
		os.Getenv("OVIRT_URL"),
		os.Getenv("OVIRT_INSECURE") != "" && os.Getenv("OVIRT_INSECURE") != "0",
		os.Getenv("OVIRT_CAFILE"),
		os.Getenv("OVIRT_CA_BUNDLE"),
		os.Getenv("OVIRT_TEST_CLUSTER_ID"),
		os.Getenv("OVIRT_BLANK_TEMPLATE_ID"),
		os.Getenv("OVIRT_TEST_STORAGEDOMAIN_ID"),
		os.Getenv("OVIRT_TEST_IMAGE_PATH"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return suite
}

func NewOvirtTestSuite(
	t *testing.T,
	ovirtUsername string,
	ovirtPassword string,
	ovirtURL string,
	ovirtInsecure bool,
	ovirtCAFile string,
	ovirtCABundle string,
	ovirtTestClusterID string,
	ovirtBlankTemplateID string,
	testStorageDomainID string,
	testImageSourcePath string,
) (suite OvirtTestSuite, err error) {
	if ovirtUsername == "" {
		return nil, fmt.Errorf("OVIRT_USERNAME not set for test case")
	}
	if ovirtPassword == "" {
		return nil, fmt.Errorf("OVIRT_PASSWORD not set for test case")
	}
	if ovirtURL == "" {
		return nil, fmt.Errorf("OVIRT_URL not set for test case")
	}
	if !ovirtInsecure && ovirtCAFile == "" && ovirtCABundle == "" {
		return nil, fmt.Errorf("OVIRT_INSECURE, OVIRT_CAFILE, or OVIRT_CA_BUNDLE must be set for acceptance tests")
	}

	provider := ovirt.ProviderContext()().(*schema.Provider)
	if err := provider.Configure(
		&terraform.ResourceConfig{
			Config: map[string]interface{}{
				"username":  ovirtUsername,
				"password":  ovirtPassword,
				"url":       ovirtURL,
				"insecure":  ovirtInsecure,
				"cafile":    ovirtCAFile,
				"ca_bundle": ovirtCABundle,
			},
		},
	); err != nil {
		return nil, fmt.Errorf("failed to configure oVirt Terraform provider (%w)", err)
	}
	providers := map[string]terraform.ResourceProvider{
		"ovirt": provider,
	}
	conn := provider.Meta().(govirt.ClientWithLegacySupport).GetSDKClient()

	if ovirtTestClusterID == "" {
		ovirtTestClusterID, err = findTestClusterID(conn)
		if err != nil {
			return nil, err
		}
	}

	if ovirtBlankTemplateID == "" {
		ovirtBlankTemplateID, err = findBlankTemplateID(conn)
		if err != nil {
			return nil, err
		}
	}

	if testStorageDomainID == "" {
		testStorageDomainID, err = findStorageDomainID(conn)
		if err != nil {
			return nil, err
		}
	}

	storageDomain, err := findStorageDomain(conn, testStorageDomainID)
	if err != nil {
		return nil, err
	}

	testCluster, err := getTestCluster(conn, ovirtTestClusterID)
	if err != nil {
		return nil, err
	}

	if testImageSourcePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("can't fetch current working directory (%w)", err)
		}
		testImageSourcePath = path.Join(
			strings.TrimSuffix(strings.TrimSuffix(cwd, `/ovirt`), `\ovirt`),
			"ovirt",
			"testimage",
			"image",
		)
	}
	if _, err := os.Stat(testImageSourcePath); err != nil {
		return nil, fmt.Errorf("test image not found in %s (%w)", testImageSourcePath, err)
	}

	hostList, err := findHosts(conn)
	if err != nil {
		return nil, fmt.Errorf("cannot list hosts (%w)", err)
	}

	testAuthz, err := findTestAuthzName(conn)
	if err != nil {
		return nil, err
	}

	testDatacenterID, testDatacenterName, err := findTestDatacenter(conn)
	if err != nil {
		return nil, err
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	cli, err := govirt.New(
		ovirtURL,
		ovirtUsername,
		ovirtPassword,
		ovirtCAFile,
		[]byte(ovirtCABundle),
		ovirtInsecure,
		nil,
		govirt.NewGoTestLogger(t),
	)
	if err != nil {
		return nil, err
	}

	return &ovirtTestSuite{
		client:              cli,
		conn:                conn,
		provider:            provider,
		providers:           providers,
		ovirtCABundle:       ovirtCABundle,
		ovirtUsername:       ovirtUsername,
		ovirtPassword:       ovirtPassword,
		ovirtURL:            ovirtURL,
		ovirtInsecure:       ovirtInsecure,
		ovirtCAFile:         ovirtCAFile,
		clusterID:           ovirtTestClusterID,
		cluster:             testCluster,
		blankTemplateID:     ovirtBlankTemplateID,
		testImageSourcePath: testImageSourcePath,
		storageDomainID:     testStorageDomainID,
		storageDomain:       storageDomain,
		hostList:            hostList,
		testAuthz:           testAuthz,
		testDatacenterName:  testDatacenterName,
		rand:                rnd,
		datacenterID:        testDatacenterID,
	}, nil
}

func getTestCluster(conn *ovirtsdk4.Connection, clusterID string) (*ovirtsdk4.Cluster, error) {
	req := conn.SystemService().ClustersService().List().Query("id", clusterID).Max(1)
	resp, err := req.Send()
	if err != nil {
		return nil, err
	}
	clusters, ok := resp.Clusters()
	if !ok {
		return nil, fmt.Errorf("failed to fetch clusters")
	}
	if len(clusters.Slice()) != 1 {
		return nil, fmt.Errorf("failed to find the previously-found test cluster")
	}
	return clusters.Slice()[0], nil
}

func findHosts(conn *ovirtsdk4.Connection) ([]string, error) {
	hostListResponse, err := conn.SystemService().HostsService().List().Send()
	if err != nil {
		return nil, fmt.Errorf("failed to list hosts (%w)", err)
	}
	var result []string
	for _, host := range hostListResponse.MustHosts().Slice() {
		result = append(result, host.MustName())
	}
	return result, nil
}

func findStorageDomainID(conn *ovirtsdk4.Connection) (string, error) {
	storageDomainsResponse, err := conn.SystemService().StorageDomainsService().List().Send()
	if err != nil {
		return "", fmt.Errorf("failed to list storage domains (%w)", err)
	}
	for _, storageDomain := range storageDomainsResponse.MustStorageDomains().Slice() {
		available, ok := storageDomain.Available()
		if !ok {
			continue
		}
		if available < 2*1024*1024*1024 {
			continue
		}
		status, ok := storageDomain.Status()
		if ok {
			if status != ovirtsdk4.STORAGEDOMAINSTATUS_ACTIVE {
				continue
			}
		}
		externalStatus, ok := storageDomain.ExternalStatus()
		if ok {
			if externalStatus != ovirtsdk4.EXTERNALSTATUS_OK {
				continue
			}
		}
		if id, ok := storageDomain.Id(); ok {
			return id, nil
		}
	}
	return "", fmt.Errorf("failed to find a working storage domain for testing")
}

func findStorageDomain(conn *ovirtsdk4.Connection, id string) (*ovirtsdk4.StorageDomain, error) {
	storageDomainResponse, err := conn.SystemService().StorageDomainsService().StorageDomainService(id).Get().Send()
	if err != nil {
		return nil, err
	}
	if storageDomain, ok := storageDomainResponse.StorageDomain(); ok {
		return storageDomain, nil
	}
	return nil, fmt.Errorf("failed to find storage domain ID %s", id)
}

func findTestClusterID(conn *ovirtsdk4.Connection) (string, error) {
	clusters, err := conn.SystemService().ClustersService().List().Send()
	if err != nil {
		return "", fmt.Errorf(
			"failed to list oVirt clusters while trying to find a cluster to run test on (%w)",
			err,
		)
	}
	for _, cluster := range clusters.MustClusters().Slice() {
		hosts, err := conn.SystemService().HostsService().List().Send()
		if err != nil {
			return "", fmt.Errorf(
				"failed to list hosts in cluster %s (%w)",
				cluster.MustName(),
				err,
			)
		}
		for _, host := range hosts.MustHosts().Slice() {
			if hostCluster, ok := host.Cluster(); ok && hostCluster.MustId() == cluster.MustId() {
				if status, ok := host.Status(); ok && status == ovirtsdk4.HOSTSTATUS_UP {
					if id, ok := cluster.Id(); ok {
						return id, nil
					}
				}
			}
		}
	}
	return "", fmt.Errorf("failed to find a valid oVirt cluster with hosts to test on")
}

func findBlankTemplateID(conn *ovirtsdk4.Connection) (string, error) {
	templatesResponse, err := conn.SystemService().TemplatesService().List().Send()
	if err != nil {
		return "", fmt.Errorf("failed to list oVirt templates while trying to find a cluster to run test on (%w)", err)
	}
	for _, tpl := range templatesResponse.MustTemplates().Slice() {
		if strings.Contains(tpl.MustDescription(), "Blank template") {
			return tpl.MustId(), nil
		}
	}
	return "", fmt.Errorf("failed to find a template with the name blank for testing")
}

func findTestAuthzName(conn *ovirtsdk4.Connection) (string, error) {
	domainList, err := conn.SystemService().DomainsService().List().Max(1).Send()
	if err != nil {
		return "", err
	}

	domains, ok := domainList.Domains()
	if !ok {
		return "", fmt.Errorf("no domains in domain list response")
	}

	for _, domain := range domains.Slice() {
		domainName, ok := domain.Name()
		if ok {
			return domainName, nil
		}
	}
	return "", fmt.Errorf("no usable domain found")
}

func findTestDatacenter(conn *ovirtsdk4.Connection) (string, string, error) {
	datacenterList, err := conn.SystemService().DataCentersService().List().Max(1).Send()
	if err != nil {
		return "", "", err
	}

	datacenters, ok := datacenterList.DataCenters()
	if !ok {
		return "", "", fmt.Errorf("no datacenters in data center list response")
	}

	for _, datacenter := range datacenters.Slice() {
		datacenterName, nameOk := datacenter.Name()
		datacenterID, idOk := datacenter.Id()
		if nameOk && idOk {
			datacenterStatus, ok := datacenter.Status()
			if ok && datacenterStatus == ovirtsdk4.DATACENTERSTATUS_UP {
				return datacenterID, datacenterName, nil
			}
		}
	}
	return "", "", fmt.Errorf("no usable datacenter found")
}

type OvirtTestSuite interface {
	// Providers returns the map of Terraform providers for use in a Terraform test suite.
	Providers() map[string]terraform.ResourceProvider

	PreCheck()

	// Client returns the oVirt client library.
	Client() govirt.Client

	// ClusterID contacts the oVirt cluster and returns the cluster ID.
	ClusterID() string

	// Cluster returns the oVirt cluster for testing purposes.
	Cluster() *ovirtsdk4.Cluster

	// EnsureVM returns a Terraform test function that checks if a VM with the specified name was successfully created
	// and publishes the name into target if so. This can be used to check attributes of a certain VM. The caller MAY
	// pass a nil pointer if no info download is desired.
	EnsureVM(terraformName string, target *ovirtsdk4.Vm) resource.TestCheckFunc

	// EnsureCluster returns a Terraform test function that checks if a Cluster with the specified name was successfully created
	EnsureCluster(terraformName string, target *ovirtsdk4.Cluster) resource.TestCheckFunc

	// EnsureVMRemoved returns a Terraform test function that checks if the specified VM was removed.
	EnsureVMRemoved(vm *ovirtsdk4.Vm) resource.TestCheckFunc

	// BlankTemplateID returns the ID of the blank template that can be used for creating dummy VMs.
	BlankTemplateID() string

	// StorageDomainID returns the ID of the storage domain to create the images on.
	StorageDomainID() string

	// StorageDomain returns the test storage domain.
	StorageDomain() *ovirtsdk4.StorageDomain

	// TestImageSourcePath returns the path to the minimal test image.
	TestImageSourcePath() string

	// TestImageSourceURL returns the source URL for an image transfer for the minimal test image.
	TestImageSourceURL() string

	// GetHostCount returns the number of hosts in the oVirt cluster used for testing.
	GetHostCount() uint

	// GetMACPoolList returns a list of MAC pool names.
	GetMACPoolList() ([]string, error)

	// TestDataSource creates a test function to check a data source
	TestDataSource(s string) resource.TestCheckFunc

	// TestResourceAttrNotEqual checks if a resource attribute is not equal to the specified value.
	TestResourceAttrNotEqual(name, key string, greaterThan bool, value interface{}) resource.TestCheckFunc

	// TestClusterDestroy creates a test function that checks if a given cluster was destroyed.
	TestClusterDestroy(cluster *ovirtsdk4.Cluster) resource.TestCheckFunc

	// GetTestAuthzName returns the name of an authz that can be used for testing purposes.
	GetTestAuthzName() string

	// GetTestDatacenterName returns the name of a datacenter that can be used for testing purposes.
	GetTestDatacenterName() string

	// CreateDisk creates a disk for testing purposes.
	CreateDisk() (*ovirtsdk4.Disk, error)

	// RemoveDisk removes a previously-created disk.
	RemoveDisk(disk *ovirtsdk4.Disk) error

	// CreateTestNetwork creates a network for testing purposes.
	CreateTestNetwork() (*ovirtsdk4.Network, error)

	// DeleteTestNetwork deletes a previously-created test network.
	DeleteTestNetwork(network *ovirtsdk4.Network) error

	// GenerateRandomID generates a random ID for testing.
	GenerateRandomID(length uint) string

	// GetTestDatacenterID returns the ID of the test datacenter
	GetTestDatacenterID() string

	// CreateNicContext creates the depending resources for NIC-related tests.
	CreateNicContext(netName string, vnicProfileName string) (*NicContext, error)

	// DestroyNicContext removes the NIC depending resources after the test.
	DestroyNicContext(nicContext *NicContext) error

	// TerraformFromTemplate will take a Go template and data variables to create a Terraform code fragment
	TerraformFromTemplate(template string, data interface{}) string
}

// NicContext contains the depending resources for a NIC-related test.
type NicContext struct {
	Network     *ovirtsdk4.Network
	VnicProfile *ovirtsdk4.VnicProfile
}

type ovirtTestSuite struct {
	conn                *ovirtsdk4.Connection
	provider            *schema.Provider
	providers           map[string]terraform.ResourceProvider
	ovirtCABundle       string
	ovirtUsername       string
	ovirtPassword       string
	ovirtURL            string
	ovirtInsecure       bool
	ovirtCAFile         string
	clusterID           string
	cluster             *ovirtsdk4.Cluster
	blankTemplateID     string
	testImageSourcePath string
	storageDomainID     string
	storageDomain       *ovirtsdk4.StorageDomain
	hostList            []string
	testAuthz           string
	testDatacenterName  string
	rand                *rand.Rand
	datacenterID        string
	client              govirt.Client
}

func (o *ovirtTestSuite) Client() govirt.Client {
	return o.client
}

func (o *ovirtTestSuite) TestImageSourceURL() string {
	return fmt.Sprintf(
		"file://%s", strings.ReplaceAll(o.TestImageSourcePath(),
		"\\",
		"/"),
	)
}

func (o *ovirtTestSuite) TerraformFromTemplate(tplText string, data interface{}) string {
	tpl := template.New("tf2")
	tpl = tpl.Funcs(template.FuncMap{
		"quoteRegexp": func(input string) string { return regexp.QuoteMeta(input) },
	})
	tpl = template.Must(tpl.Parse(tplText))
	wr := &bytes.Buffer{}
	if err := tpl.Execute(wr, data); err != nil {
		panic(err)
	}
	return wr.String()
}

func (o *ovirtTestSuite) StorageDomain() *ovirtsdk4.StorageDomain {
	return o.storageDomain
}

func (o *ovirtTestSuite) CreateNicContext(netName string, vnicProfileName string) (*NicContext, error) {
	network, err := o.createNetwork(netName)
	if err != nil {
		return &NicContext{}, fmt.Errorf("failed to create network for NIC context (%w)", err)
	}
	vnicProfile, err := o.createVnicProfile(vnicProfileName, network)
	return &NicContext{
		Network:     network,
		VnicProfile: vnicProfile,
	}, err
}

func (o *ovirtTestSuite) DestroyNicContext(nicContext *NicContext) error {
	if nicContext.VnicProfile != nil {
		if err := o.removeVnicProfile(nicContext.VnicProfile); err != nil {
			return fmt.Errorf("failed to remove vnic profile (%w)", err)
		}
	}
	if nicContext.Network != nil {
		if err := o.removeNetwork(nicContext.Network); err != nil {
			return fmt.Errorf("failed to remove network (%w)", err)
		}
	}
	return nil
}

func (o *ovirtTestSuite) GetTestDatacenterID() string {
	return o.datacenterID
}

func (o *ovirtTestSuite) GenerateRandomID(length uint) string {
	return o.generateRandomID(length)
}

func (o *ovirtTestSuite) createVm(name string) (*ovirtsdk4.Vm, error) {
	vm := ovirtsdk4.NewVmBuilder().
		Cluster(ovirtsdk4.NewClusterBuilder().Id(o.clusterID).MustBuild()).
		StorageDomain(ovirtsdk4.NewStorageDomainBuilder().Id(o.storageDomainID).MustBuild()).
		Name(name).
		Template(ovirtsdk4.NewTemplateBuilder().Name(o.blankTemplateID).MustBuild()).
		MustBuild()
	result, err := o.conn.SystemService().VmsService().Add().Vm(vm).Send()
	if err != nil {
		return nil, err
	}
	return result.MustVm(), nil
}

func (o *ovirtTestSuite) createVnicProfile(name string, network *ovirtsdk4.Network) (*ovirtsdk4.VnicProfile, error) {
	profile := ovirtsdk4.NewVnicProfileBuilder().Name(name).Network(network).MustBuild()
	result, err := o.conn.SystemService().VnicProfilesService().Add().Profile(profile).Send()
	if err != nil {
		return nil, err
	}
	return result.MustProfile(), nil
}

func (o *ovirtTestSuite) createNetwork(name string) (*ovirtsdk4.Network, error) {
	net, err := ovirtsdk4.NewNetworkBuilder().
		Name(name).
		DataCenter(ovirtsdk4.NewDataCenterBuilder().Name(o.testDatacenterName).MustBuild()).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build network to be added (%w)", err)
	}
	response, err := o.conn.SystemService().NetworksService().Add().Network(net).Send()
	if err != nil {
		return nil, fmt.Errorf("failed to create test network (%w)", err)
	}
	network, ok := response.Network()
	if !ok {
		return nil, fmt.Errorf("no network returned when creating test network")
	}
	if _, err := o.conn.SystemService().ClustersService().ClusterService(o.clusterID).NetworksService().Add().Network(network).Send(); err != nil {
		return network, fmt.Errorf("failed to attach network to cluster (%w)", err)
	}
	return network, nil
}

func (o *ovirtTestSuite) CreateTestNetwork() (*ovirtsdk4.Network, error) {
	netName := fmt.Sprintf("terraform-test-%s", o.generateRandomID(5))
	return o.createNetwork(netName)
}

func (o *ovirtTestSuite) DeleteTestNetwork(network *ovirtsdk4.Network) (err error) {
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	for {
		_, err = o.conn.SystemService().NetworksService().NetworkService(network.MustId()).Remove().Send()
		if err == nil {
			return
		}
		select {
		case <-timeout.Done():
			return fmt.Errorf("timeout (%w)", err)
		case <-time.After(10 * time.Second):
		}
	}
}

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (o *ovirtTestSuite) generateRandomID(n uint) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[o.rand.Intn(len(letters))]
	}
	return string(b)
}

func (o *ovirtTestSuite) CreateDisk() (disk *ovirtsdk4.Disk, err error) {
	diskName := fmt.Sprintf("test_disk_%s", o.generateRandomID(5))
	disk, err = ovirtsdk4.NewDiskBuilder().
		Name(diskName).
		Format(ovirtsdk4.DISKFORMAT_RAW).
		ProvisionedSize(1024 * 1024).
		StorageDomainsOfAny(
			ovirtsdk4.NewStorageDomainBuilder().
				Id(o.storageDomainID).
				MustBuild(),
		).
		Build()
	if err != nil {
		// This should never happen
		panic(err)
	}
	diskResponse, err := o.conn.SystemService().DisksService().Add().Disk(disk).Send()
	if err != nil {
		return nil, fmt.Errorf("failed to create disk (%w)", err)
	}
	disk, ok := diskResponse.Disk()
	if !ok {
		return nil, fmt.Errorf("no disk object returned from disk creation")
	}

	timeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancelFunc()
	for {
		diskResponse, err := o.conn.SystemService().DisksService().DiskService(disk.MustId()).Get().Send()
		if err == nil {
			if disk, ok := diskResponse.Disk(); ok {
				if diskStatus, ok := disk.Status(); ok {
					if diskStatus == ovirtsdk4.DISKSTATUS_OK {
						return disk, nil
					}
				}
			}
		}
		select {
		case <-timeout.Done():
			return disk, fmt.Errorf("timeout while waiting for disk to come up")
		case <-time.After(10 * time.Second):
		}
	}
}

func (o *ovirtTestSuite) RemoveDisk(disk *ovirtsdk4.Disk) (err error) {
	_, err = o.conn.SystemService().DisksService().DiskService(disk.MustId()).Remove().Send()
	if err != nil {
		return fmt.Errorf("failed to remove disk (%w)", err)
	}
	return nil
}

func (o *ovirtTestSuite) GetTestDatacenterName() string {
	return o.testDatacenterName
}

func (o *ovirtTestSuite) GetTestAuthzName() string {
	return o.testAuthz
}

func (o *ovirtTestSuite) TestClusterDestroy(cluster *ovirtsdk4.Cluster) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		listResponse,err := o.conn.SystemService().ClustersService().List().Search("name="+cluster.MustName()).Send()

		if err != nil {
			return err
		}

		if len(listResponse.MustClusters().Slice()) != 0 {
			return fmt.Errorf("cluster %s has not been removed after test", cluster.MustName())
		}
		return nil
	}
}

func (o *ovirtTestSuite) TestResourceAttrNotEqual(
	name, key string,
	greaterThan bool,
	value interface{},
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[name]
		v, ok := rs.Primary.Attributes[key]
		if !ok {
			return fmt.Errorf("%s: Attribute '%s' not found", name, key)
		}

		valueV := reflect.ValueOf(value)

		var valueString string

		switch valueV.Kind() {
		case reflect.Bool:
			return fmt.Errorf("for bool type, please use `resource.TestCheckResourceAttr` instead")
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			valueString = fmt.Sprintf("%d", value)
		case reflect.String:
			valueString = fmt.Sprintf("%s", value)
		case reflect.Float32, reflect.Float64:
			valueString = fmt.Sprintf("%f", value)
		default:
			return fmt.Errorf("attr not equal check only supports int/int32/int64/float/float64/string")
		}

		var firstOptLabel, secondOptLabel string
		if greaterThan {
			firstOptLabel = ">"
			secondOptLabel = "<"
		} else {
			firstOptLabel = "<"
			secondOptLabel = ">"
		}

		if v > valueString != greaterThan {
			return fmt.Errorf(
				"%[1]s: Attribute '%[2]s' expected %#[3]v %[5]s %#[4]v, got %#[3]v %[6]s %#[4]v",
				name,
				key,
				v,
				valueString,
				firstOptLabel,
				secondOptLabel)
		}
		return nil
	}
}

func (o *ovirtTestSuite) TestDataSource(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find data source: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("data source ID not set")
		}
		return nil
	}
}

func (o *ovirtTestSuite) Cluster() *ovirtsdk4.Cluster {
	return o.cluster
}

func (o *ovirtTestSuite) GetMACPoolList() ([]string, error) {
	macPoolsResponse, err := o.conn.SystemService().MacPoolsService().List().Send()
	if err != nil {
		return nil, fmt.Errorf("failed to list MAC pools (%w)", err)
	}
	var result []string
	for _, pool := range macPoolsResponse.MustPools().Slice() {
		result = append(result, pool.MustName())
	}
	return result, nil
}

func (o *ovirtTestSuite) GetHostCount() uint {
	return uint(len(o.hostList))
}

func (o *ovirtTestSuite) StorageDomainID() string {
	return o.storageDomainID
}

func (o *ovirtTestSuite) TestImageSourcePath() string {
	return o.testImageSourcePath
}

func (o *ovirtTestSuite) BlankTemplateID() string {
	return o.blankTemplateID
}

func (o *ovirtTestSuite) PreCheck() {

}

func (o *ovirtTestSuite) ClusterID() string {
	return o.clusterID
}

func (o *ovirtTestSuite) Providers() map[string]terraform.ResourceProvider {
	return o.providers
}

func (o *ovirtTestSuite) EnsureCluster(terraformName string, cl *ovirtsdk4.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[terraformName]
		if !ok {
			return fmt.Errorf("not found: %s", terraformName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no VM ID is set")
		}

		getResp, err := o.conn.SystemService().ClustersService().
			ClusterService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return err
		}
		cluster, ok := getResp.Cluster()
		if ok {
			*cl = *cluster
			return nil
		}
		return fmt.Errorf("cluster %s not exist", rs.Primary.ID)
	}
}

func (o *ovirtTestSuite) EnsureVM(terraformName string, vm *ovirtsdk4.Vm) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[terraformName]
		if !ok {
			return fmt.Errorf("not found: %s", terraformName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no VM ID is set")
		}
		getResp, err := o.conn.SystemService().VmsService().
			VmService(rs.Primary.ID).
			Get().
			Send()
		if err != nil {
			return err
		}
		realVM, ok := getResp.Vm()
		if ok {
			*vm = *realVM
			return nil
		}
		return fmt.Errorf("VM %s not exist", rs.Primary.ID)
	}
}

func (o *ovirtTestSuite) EnsureVMRemoved(vm *ovirtsdk4.Vm) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, ok := vm.Id(); !ok {
			return nil
		}
		_, err := o.conn.SystemService().VmsService().
			VmService(vm.MustId()).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				return nil
			}
			return err
		}
		return fmt.Errorf("VM %s is not removed", vm.MustId())
	}
}

func (o *ovirtTestSuite) removeNetwork(network *ovirtsdk4.Network) error {
	_, err := o.conn.SystemService().NetworksService().NetworkService(network.MustId()).Remove().Send()
	return err
}

func (o *ovirtTestSuite) removeVnicProfile(profile *ovirtsdk4.VnicProfile) error {
	_, err := o.conn.SystemService().VnicProfilesService().ProfileService(profile.MustId()).Remove().Send()
	return err
}

func (o *ovirtTestSuite) removeVm(vm *ovirtsdk4.Vm) error {
	_, err := o.conn.SystemService().VmsService().VmService(vm.MustId()).Remove().Send()
	return err
}
