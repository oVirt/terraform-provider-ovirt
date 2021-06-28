package ovirtclient

import (
	"fmt"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

// ClusterClient is a part of the Client that deals with clusters in the oVirt Engine.
type ClusterClient interface {
	// ListClusters returns a list of all clusters in the oVirt engine.
	ListClusters() ([]Cluster, error)
	// GetCluster returns a specific cluster based on the cluster ID. An error is returned if the cluster doesn't exist.
	GetCluster(id string) (Cluster, error)
}

// Cluster represents a cluster returned from a ListClusters or GetCluster call.
type Cluster interface {
	// ID returns the UUID of the cluster.
	ID() string
	// Name returns the textual name of the cluster.
	Name() string
}

func convertSDKCluster(sdkCluster *ovirtsdk4.Cluster) (Cluster, error) {
	id, ok := sdkCluster.Id()
	if !ok {
		return nil, fmt.Errorf("failed to fetch ID for cluster")
	}

	name, ok := sdkCluster.Name()
	if !ok {
		return nil, fmt.Errorf("failed to fetch name for cluster %s", id)
	}
	return &cluster{
		id:   id,
		name: name,
	}, nil
}

type cluster struct {
	id   string
	name string
}

func (c cluster) ID() string {
	return c.id
}

func (c cluster) Name() string {
	return c.name
}
