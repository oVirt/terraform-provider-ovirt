package govirt

import (
	"fmt"
)

func (o *oVirtClient) ListTemplates() ([]Template, error) {
	response, err := o.conn.SystemService().TemplatesService().List().Send()
	if err != nil {
		return nil, fmt.Errorf("failed to list templates (%w)", err)
	}
	sdkTemplates, ok := response.Templates()
	if !ok {
		return nil, fmt.Errorf("host list response didn't contain hosts")
	}
	result := make([]Template, len(sdkTemplates.Slice()))
	for i, sdkTemplate := range sdkTemplates.Slice() {
		result[i], err = convertSDKTemplate(sdkTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to convert host %d in listing (%w)", i, err)
		}
	}
	return result, nil
}
