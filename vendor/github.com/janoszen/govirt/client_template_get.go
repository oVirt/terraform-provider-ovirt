package govirt

import (
	"fmt"
)

func (o *oVirtClient) GetTemplate(id string) (Template, error) {
	response, err := o.conn.SystemService().TemplatesService().TemplateService(id).Get().Send()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch template %s (%w)", id, err)
	}
	sdkTemplate, ok := response.Template()
	if !ok {
		return nil, fmt.Errorf("API response contained no host")
	}
	template, err := convertSDKTemplate(sdkTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to convert template object (%w)", err)
	}
	return template, nil
}
