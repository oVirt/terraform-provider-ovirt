package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) GetHost(id string) (Host, error) {
	response, err := o.conn.SystemService().HostsService().HostService(id).Get().Send()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch host %s (%w)", id, err)
	}
	sdkHost, ok := response.Host()
	if !ok {
		return nil, fmt.Errorf("API response contained no host")
	}
	host, err := convertSDKHost(sdkHost)
	if err != nil {
		return nil, fmt.Errorf("failed to convert host object (%w)", err)
	}
	return host, nil
}
