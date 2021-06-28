package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) GetCluster(id string) (cluster Cluster, err error) {
	response, err := o.conn.SystemService().ClustersService().ClusterService(id).Get().Send()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cluster ID %s (%w)", id, err)
	}
	sdkCluster, ok := response.Cluster()
	if !ok {
		return nil, fmt.Errorf("no cluster returned when getting cluster ID %s", id)
	}
	cluster, err = convertSDKCluster(sdkCluster)
	if err != nil {
		return nil, fmt.Errorf("failed to convert cluster %s (%w)", id, err)
	}
	return cluster, nil
}
