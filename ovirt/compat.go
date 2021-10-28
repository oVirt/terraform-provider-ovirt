package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func validateCompat(v func(i interface{}, path cty.Path) diag.Diagnostics) schema.SchemaValidateFunc {
	return func(i interface{}, s string) (warnings []string, errors []error) {
		results := v(i, cty.GetAttrPath(s))

		for _, result := range results {
			if result.Severity == diag.Error {
				errors = append(errors, fmt.Errorf("%s (%s)", result.Summary, result.Detail))
			} else {
				warnings = append(warnings, fmt.Sprintf("%s\n%s", result.Summary, result.Detail))
			}
		}
		return
	}
}

func crudCompat(
	f func(ctx context.Context, data *schema.ResourceData, d interface{}) diag.Diagnostics,
) func(*schema.ResourceData, interface{}) error {
	return func(data *schema.ResourceData, i interface{}) error {
		return diagsToError(f(context.Background(), data, i))
	}
}

func importCompat(
	importF func(ctx context.Context, data *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error),
) schema.StateFunc {
	return func(data *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
		return importF(context.Background(), data, i)
	}
}
