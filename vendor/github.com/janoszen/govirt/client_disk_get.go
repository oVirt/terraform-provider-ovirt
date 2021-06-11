package govirt

import (
	"fmt"
)

func (o *oVirtClient) GetDisk(diskID string) (Disk, error) {
	response, err := o.conn.SystemService().DisksService().DiskService(diskID).Get().Send()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch disk %s (%w)", diskID, err)
	}
	sdkDisk, ok := response.Disk()
	if !ok {
		return nil, fmt.Errorf("disk %s response did not contain a disk (%w)", diskID, err)
	}
	disk, err := convertSDKDisk(sdkDisk)
	if err != nil {
		return nil, fmt.Errorf("failed to convert disk %s (%w)", diskID, err)
	}
	return disk, nil
}
