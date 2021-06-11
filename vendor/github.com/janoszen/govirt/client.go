package govirt

import (
	"net/http"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

// Client is a simplified client for the oVirt API.
type Client interface {
	// GetSDKClient returns a configured oVirt SDK client for the use cases that are not covered by goVirt.
	GetSDKClient() *ovirtsdk4.Connection

	// GetHTTPClient returns a configured HTTP client for the oVirt engine. This can be used to send manual
	// HTTP request to the oVirt engine.
	GetHTTPClient() http.Client

	// GetURL returns the oVirt engine base URL.
	GetURL() string

	DiskClient
	VMClient
	ClusterClient
	StorageDomainClient
	HostClient
	TemplateClient
}
