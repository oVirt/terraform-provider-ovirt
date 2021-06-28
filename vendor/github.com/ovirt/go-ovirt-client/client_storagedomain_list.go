package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) ListStorageDomains() (storageDomains []StorageDomain, err error) {
	response, err := o.conn.SystemService().StorageDomainsService().List().Send()
	if err != nil {
		return nil, fmt.Errorf("failed to list storage domains (%w)", err)
	}
	sdkStorageDomains, ok := response.StorageDomains()
	if !ok {
		return nil, fmt.Errorf("API call did not return storage domains in response")
	}
	storageDomains = make([]StorageDomain, len(sdkStorageDomains.Slice()))
	for i, sdkStorageDomain := range sdkStorageDomains.Slice() {
		storageDomain, err := convertSDKStorageDomain(sdkStorageDomain)
		if err != nil {
			return nil, fmt.Errorf("failed to convert storage domain %d in listing (%w)", i, err)
		}
		storageDomains[i] = storageDomain
	}
	return storageDomains, nil
}
