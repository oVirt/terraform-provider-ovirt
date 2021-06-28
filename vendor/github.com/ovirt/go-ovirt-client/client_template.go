package ovirtclient

import (
	"fmt"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

type TemplateClient interface {
	ListTemplates() ([]Template, error)
	GetTemplate(id string) (Template, error)
}

type Template interface {
	ID() string
	Name() string
	Description() string
}

func convertSDKTemplate(sdkTemplate *ovirtsdk4.Template) (Template, error) {
	id, ok := sdkTemplate.Id()
	if !ok {
		return nil, fmt.Errorf("template does not contain ID")
	}
	name, ok := sdkTemplate.Name()
	if !ok {
		return nil, fmt.Errorf("template does not contain a name")
	}
	description, ok := sdkTemplate.Description()
	if !ok {
		return nil, fmt.Errorf("template does not contain a description")
	}
	return &template{
		id:          id,
		name:        name,
		description: description,
	}, nil
}

type template struct {
	id          string
	name        string
	description string
}

func (t template) ID() string {
	return t.id
}

func (t template) Name() string {
	return t.name
}

func (t template) Description() string {
	return t.description
}
