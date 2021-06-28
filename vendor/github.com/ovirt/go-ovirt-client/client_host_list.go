package ovirtclient

import (
	"fmt"
)

func (o *oVirtClient) ListHosts() ([]Host, error) {
	response, err := o.conn.SystemService().HostsService().List().Send()
	if err != nil {
		return nil, fmt.Errorf("failed to list hosts (%w)", err)
	}
	sdkHosts, ok := response.Hosts()
	if !ok {
		return nil, fmt.Errorf("host list response didn't contain hosts")
	}
	result := make([]Host, len(sdkHosts.Slice()))
	for i, sdkHost := range sdkHosts.Slice() {
		result[i], err = convertSDKHost(sdkHost)
		if err != nil {
			return nil, fmt.Errorf("failed to convert host %d in listing (%w)", i, err)
		}
	}
	return result, nil
}
