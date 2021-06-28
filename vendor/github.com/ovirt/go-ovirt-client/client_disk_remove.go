package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) RemoveDisk(diskID string) error {
	if _, err := o.conn.SystemService().DisksService().DiskService(diskID).Remove().Send(); err != nil {
		return fmt.Errorf("failed to remove disk %s (%w)", diskID, err)
	}
	return nil
}
