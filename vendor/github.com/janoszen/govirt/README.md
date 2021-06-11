# goVirt: an easy-to-use overlay for the oVirt Go SDK

<p align="center"><strong>⚠⚠⚠ This library is work-in-progress. Do not use. ⚠⚠⚠</strong></p>

This library provides an easy-to-use overlay for the automatically generated [Go SDK for oVirt](https://github.com/oVirt/go-ovirt). It does *not* replace the Go SDK. It implements the functions of the SDK only partially and is primarily used by the [oVirt Terraform provider](https://github.com/oVirt/terraform-provider-ovirt/).

## Using this library

To use this library you will have to include it as a Go module dependency:

```
go get github.com/janoszen/govirt
```

You can then create a client instance like this:

```go
package main

import "github.com/janoszen/govirt"

func main() {
    // Create a logger that logs to the standard Go log here:
    logger := govirt.NewGoLogLogger()
    // Create a new goVirt instance:
	client, err := govirt.New(
        // URL to your oVirt engine API here:
        "https://your-ovirt-engine/ovirt-engine/api/",
        // Username here:
        "admin@internal",
        // Password here:
        "password-here",
        // Provide the path to the CA certificate here:
        "/path/to/ca.crt",
        // Alternatively, provide the certificate directly:
        []byte("ca-cert-here in PEM format"),
        // Disable certificate verification. This is a bad idea:
        false,
        // Extra headers map:
        map[string]string{},
        logger,
    )
    if err != nil {
        // Handle error, here in a really crude way:
    	panic(err)
    }
    // Use client. Please use the code completion in your IDE to
    // discover the functions. Each is well documented.
    upload, err := client.StartImageUpload(
        //...
    )
    //....
}
```

## Test helper

This library also provides a test helper for integration testing against the oVirt engine. It allows for automatically discovering a usable storage domain, host, clusters, etc:

```go
package main

import (
  "os"
  "testing"

  "github.com/janoszen/govirt"
)

func TestSomething(t *testing.T) {
    // Create a logger that logs to the standard Go log here:
    logger := govirt.NewGoTestLogger(t)
    // Create the test helper
    helper, err := govirt.NewTestHelper(
        os.Getenv("OVIRT_URL"),
        os.Getenv("OVIRT_USER"),
        os.Getenv("OVIRT_PASSWORD"),
        os.Getenv("OVIRT_CAFILE"),
        []byte(os.Getenv("OVIRT_CABUNDLE")),
        os.Getenv("OVIRT_INSECURE") != "",
        os.Getenv("OVIRT_CLUSTER_ID"),
        os.Getenv("OVIRT_BLANK_TEMPLATE_ID"),
        os.Getenv("OVIRT_STORAGE_DOMAIN_ID"),
        logger,
    )
    if err != nil {
        t.Fatal(err)
    }
    // Fetch the cluster ID for testing
    clusterID := helper.GetClusterID()
    //...
}
```

## FAQ

### Why doesn't the library return the underlying oVirt SDK objects?

It's a painful decision we made. We want to encourage anyone who needs a certain function to submit a PR instead of simply relying on the SDK objects. This will lead to some overhead when a new function needs to be added, but leads to cleaner code in the end and makes this library more comprehensive.