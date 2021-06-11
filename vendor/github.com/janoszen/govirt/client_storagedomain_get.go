package govirt

import (
	"fmt"
)

func (o *oVirtClient) GetStorageDomain(id string) (storageDomain StorageDomain, err error) {
	response, err := o.conn.SystemService().StorageDomainsService().StorageDomainService(id).Get().Send()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage domain %s (%w)", id, err)
	}
	sdkStorageDomain, ok := response.StorageDomain()
	if !ok {
		return nil, fmt.Errorf("response did not contain a storage domain")
	}
	storageDomain, err = convertSDKStorageDomain(sdkStorageDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to convert storage domain (%w)", err)
	}
	return storageDomain, nil
}
