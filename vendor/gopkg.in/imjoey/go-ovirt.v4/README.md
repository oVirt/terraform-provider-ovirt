# oVirt Go SDK [![Build Status](https://travis-ci.org/imjoey/ovirt-engine-sdk-go.svg?branch=master)](https://travis-ci.org/imjoey/ovirt-engine-sdk-go)

## Introduction

The oVirt Go SDK is a Go package that simplyfies access to the
oVirt Engine API.

> __IMPORTANT__: The code in this project is generated automatically by the [imjoey/ovirt-engine-sdk-go](https://github.com/imjoey/ovirt-engine-sdk-go). So if you want to know how to generate the code, please read the `README.md` in the  [imjoey/ovirt-engine-sdk-go](https://github.com/imjoey/ovirt-engine-sdk-go) repositorys instead.

## Usage

To use the SDK you should import ovirtsdk package as follows:

```go
import (
    ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4" // v4 <-> ovirt 4.x
)
```

> __IMPORTANT__: In product envrionment, you should __NEVER__ use `import "github.com/imjoey/go-ovirt"` that it imports the master branch which will always be under heavy development.

That will give you access to all the classes of the SDK, and in particular
to the `Connection` class. This is the entry point of the SDK,
and gives you access to the root of the tree of services of the API:

```go
import (
    "fmt"
    "time"
    ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4"
)

// Create the connection to the api server
inputRawURL := "https://10.1.111.229/ovirt-engine/api"
conn, err := ovirtsdk4.NewConnectionBuilder().
	URL(inputRawURL).
	Username("admin@internal").
	Password("qwer1234").
	Insecure(true).
	Compress(true).
	Timeout(time.Second * 10).
	Build()
if err != nil {
	t.Fatalf("Make connection failed, reason: %s", err.Error())
}

// Never forget to close connection
defer conn.Close()

```

There are two ways of using the SDK, one is calling regular functions, which should check if error returned, the other is calling functions prefixed with `Must`, which is short and chain-function calling supported.

Calling the regular functions is recommended, because it is  more accurate for catching errors.

> __IMPORTANT__: you should catch the panic errors by defining the recover-defer function.


### Regular _Recommended_

```go
import (
    "fmt"
    "time"
    ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4"
)

// Create the connection to the api server
inputRawURL := "https://10.1.111.229/ovirt-engine/api"
conn, err := ovirtsdk4.NewConnectionBuilder().
	URL(inputRawURL).
	Username("admin@internal").
	Password("qwer1234").
	Insecure(true).
	Compress(true).
	Timeout(time.Second * 10).
	Build()
if err != nil {
	t.Fatalf("Make connection failed, reason: %s", err.Error())
}

defer conn.Close()

// Get the reference to the "clusters" service
clustersService := conn.SystemService().ClustersService()

// Use the "list" method of the "clusters" service to list all the clusters of the system
clustersResponse, err := clustersService.List().Send()
if err != nil {
	fmt.Printf("Failed to get cluster list, reason: %v\n", err)
	return
}

if clusters, ok := clustersResponse.Clusters(); ok {
	// Print the datacenter names and identifiers
	fmt.Printf("Cluster: (")
	for _, cluster := range clusters.Slice() {
		if clusterName, ok := cluster.Name(); ok {
			fmt.Printf(" name: %v", clusterName)
		}
		if clusterId, ok := cluster.Id(); ok {
			fmt.Printf(" id: %v", clusterId)
		}
	}
	fmt.Println(")")
}

```

### Must _Not-Recommended_

```go
import (
    "fmt"
    "time"
    ovirtsdk4 "gopkg.in/imjoey/go-ovirt.v4"
)

// Create the connection to the api server
inputRawURL := "https://10.1.111.229/ovirt-engine/api"
conn, err := ovirtsdk4.NewConnectionBuilder().
	URL(inputRawURL).
	Username("admin@internal").
	Password("qwer1234").
	Insecure(true).
	Compress(true).
	Timeout(time.Second * 10).
	Build()
if err != nil {
	t.Fatalf("Make connection failed, reason: %s", err.Error())
}

defer conn.Close()

// To use `Must` methods, you should recover it if panics
defer func() {
	if err := recover(); err != nil {
		fmt.Printf("Panics occurs %v, try the non-Must methods to find the reason", err)
	}
}()

// Get the reference to the "clusters" service
clustersService := conn.SystemService().ClustersService()

// Use the "list" method of the "clusters" service to list all the clusters of the system
clustersResponse := clustersService.List().MustSend()

clusters := clustersResponse.MustClusters()

// Print the datacenter names and identifiers
fmt.Printf("Cluster: (")
for _, cluster := range clusters.Slice() {
	fmt.Printf(" name: %v", cluster.MustName())
	fmt.Printf(" id: %v", cluster.MustId())
}
fmt.Println(")")

```

## More examples

You could refer to more examples under `examples/` directory.