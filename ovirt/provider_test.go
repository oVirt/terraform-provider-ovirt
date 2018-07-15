package ovirt

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"ovirt": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("OVIRT_USERNAME"); v == "" {
		t.Fatal("OVIRT_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("OVIRT_PASSWORD"); v == "" {
		t.Fatal("OVIRT_PASSWORD must be set for acceptance tests")
	}
	if v := os.Getenv("OVIRT_URL"); v == "" {
		t.Fatal("OVIRT_URL must be set for acceptance tests")
	}
}

func testAccCheckOvirtDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find data source: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("data source ID not set")
		}
		return nil
	}
}

func testCheckResourceAttrNotEqual(name, key string, greaterThan bool, value interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[name]
		v, ok := rs.Primary.Attributes[key]
		if !ok {
			return fmt.Errorf("%s: Attribute '%s' not found", name, key)
		}

		valueV := reflect.ValueOf(value)

		var valueString string

		switch valueV.Kind() {
		case reflect.Bool:
			return fmt.Errorf("for bool type, please use `resource.TestCheckResourceAttr` instead")
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			valueString = fmt.Sprintf("%d", value)
		case reflect.String:
			valueString = fmt.Sprintf("%s", value)
		case reflect.Float32, reflect.Float64:
			valueString = fmt.Sprintf("%f", value)
		default:
			return fmt.Errorf("attr not equal check only supports int/int32/int64/float/float64/string")
		}

		var firstOptLabel, secondOptLabel string
		if greaterThan {
			firstOptLabel = ">"
			secondOptLabel = "<"
		} else {
			firstOptLabel = "<"
			secondOptLabel = ">"
		}

		if v > valueString != greaterThan {
			return fmt.Errorf(
				"%[1]s: Attribute '%[2]s' expected %#[3]v %[5]s %#[4]v, got %#[3]v %[6]s %#[4]v",
				name,
				key,
				v,
				valueString,
				firstOptLabel,
				secondOptLabel)
		}
		return nil
	}
}
