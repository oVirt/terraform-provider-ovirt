package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) ListClusters() ([]Cluster, error) {
	clustersResponse, err := o.conn.SystemService().ClustersService().List().Send()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to list oVirt clusters (%w)",
			err,
		)
	}
	sdkClusters, ok := clustersResponse.Clusters()
	if !ok {
		return nil, fmt.Errorf("no clusters returned from clusters list API call")
	}
	clusters := make([]Cluster, len(sdkClusters.Slice()))
	for i, sdkCluster := range sdkClusters.Slice() {
		clusters[i], err = convertSDKCluster(sdkCluster)
		if err != nil {
			return nil, fmt.Errorf("failed to convert cluster during cluster listing item %d (%w)", i, err)
		}
	}
	return clusters, nil
}
