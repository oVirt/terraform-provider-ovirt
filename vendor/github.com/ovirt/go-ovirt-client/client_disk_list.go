package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) ListDisks() ([]Disk, error) {
	response, err := o.conn.SystemService().DisksService().List().Send()
	if err != nil {
		return nil, fmt.Errorf("failed to list disks (%w)", err)
	}
	sdkDisks, ok := response.Disks()
	if !ok {
		return nil, fmt.Errorf("disk list response does not contain disks")
	}
	result := make([]Disk, len(sdkDisks.Slice()))
	for i, sdkDisk := range sdkDisks.Slice() {
		disk, err := convertSDKDisk(sdkDisk)
		if err != nil {
			return nil, fmt.Errorf("failed to convert disk %d (%w)", i, err)
		}
		result[i] = disk
	}
	return result, nil
}
