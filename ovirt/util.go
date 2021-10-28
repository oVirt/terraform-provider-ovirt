package ovirt

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	ovirtclient "github.com/ovirt/go-ovirt-client"
)

func extractString(data *schema.ResourceData, option string, diags diag.Diagnostics) (string, diag.Diagnostics) {
	var url string
	urlInterface, ok := data.GetOk(option)
	if !ok {
		diags = append(
			diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("The %s option is not set", option),
				Detail:   fmt.Sprintf("The %s option must be set if mock=false", option),
			},
		)
	} else {
		url, ok = urlInterface.(string)
		if !ok {
			diags = append(
				diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("The %s option is not a string", option),
					Detail:   fmt.Sprintf("The %s option must be set and be a string", option),
				},
			)
		}
	}
	return url, diags
}

func setResourceField(data *schema.ResourceData, field string, value interface{}, diags diag.Diagnostics) diag.Diagnostics {
	if err := data.Set(field, value); err != nil {
		diags = append(
			diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to update %s field", field),
				Detail:   err.Error(),
			},
		)
	}
	return diags
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	var e ovirtclient.EngineError
	if errors.As(err, &e) {
		return e.HasCode(ovirtclient.ENotFound)
	}
	return false
}

func diagsToError(diags diag.Diagnostics) error {
	if !diags.HasError() {
		return nil
	}
	errs := make([]string, len(diags))
	for i, d := range diags {
		errs[i] = fmt.Sprintf("%s (%s)", d.Summary, d.Detail)
	}
	return fmt.Errorf("%s", strings.Join(errs, ", "))
}

func errorToDiags(action string, err error) diag.Diagnostics {
	if err != nil {
		return diag.Diagnostics{errorToDiag(action, err)}
	}
	return nil
}

func errorToDiag(action string, err error) diag.Diagnostic {
	if err == nil {
		panic(fmt.Errorf("errorToDiag received nil error"))
	}
	return diag.Diagnostic{
		Severity: diag.Error,
		Summary:  fmt.Sprintf("Failed to %s", action),
		Detail:   err.Error(),
	}
}
