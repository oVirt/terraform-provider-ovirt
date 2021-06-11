package govirt

import (
	"context"
)

func (o *oVirtClient) CreateVM(
	ctx context.Context,
	clusterID string,
	cpuTopo VMCPUTopo,
	templateID string,
	blockDevices []VMBlockDevice,
) {
	panic("implement me")
}
