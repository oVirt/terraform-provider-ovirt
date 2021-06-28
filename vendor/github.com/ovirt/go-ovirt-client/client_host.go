package ovirtclient

import (
	"fmt"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

// HostClient contains the API portion that deals with hosts.
type HostClient interface {
	ListHosts() ([]Host, error)
	GetHost(id string) (Host, error)
}

// Host is the representation of a host returned from the oVirt Engine API.
type Host interface {
	ID() string
	ClusterID() string
	Status() HostStatus
}

type HostStatus string

const (
	HostStatusConnecting              HostStatus = "connecting"
	HostStatusDown                    HostStatus = "down"
	HostStatusError                   HostStatus = "error"
	HostStatusInitializing            HostStatus = "initializing"
	HostStatusInstallFailed           HostStatus = "install_failed"
	HostStatusInstalling              HostStatus = "installing"
	HostStatusInstallingOS            HostStatus = "installing_os"
	HostStatusKDumping                HostStatus = "kdumping"
	HostStatusMaintenance             HostStatus = "maintenance"
	HostStatusNonOperational          HostStatus = "non_operational"
	HostStatusNonResponsive           HostStatus = "non_responsive"
	HostStatusPendingApproval         HostStatus = "pending_approval"
	HostStatusPreparingForMaintenance HostStatus = "preparing_for_maintenance"
	HostStatusReboot                  HostStatus = "reboot"
	HostStatusUnassigned              HostStatus = "unassigned"
	HostStatusUp                      HostStatus = "up"
)

func convertSDKHost(sdkHost *ovirtsdk4.Host) (Host, error) {
	id, ok := sdkHost.Id()
	if !ok {
		return nil, fmt.Errorf("returned host did not contain an ID")
	}
	status, ok := sdkHost.Status()
	if !ok {
		return nil, fmt.Errorf("returned host did not contain a status")
	}
	sdkCluster, ok := sdkHost.Cluster()
	if !ok {
		return nil, fmt.Errorf("returned host did not contain a cluster")
	}
	clusterID, ok := sdkCluster.Id()
	if !ok {
		return nil, fmt.Errorf("failed to fetch cluster ID from host %s", id)
	}
	return &host{
		id:        id,
		status:    HostStatus(status),
		clusterID: clusterID,
	}, nil
}

type host struct {
	id        string
	clusterID string
	status    HostStatus
}

func (h host) ID() string {
	return h.id
}

func (h host) ClusterID() string {
	return h.clusterID
}

func (h host) Status() HostStatus {
	return h.status
}
