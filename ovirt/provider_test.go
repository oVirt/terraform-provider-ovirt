// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.
package ovirt

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = ProviderContext()().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"ovirt": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := ProviderContext()().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = ProviderContext()()
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
	ovirtInsecure := os.Getenv("OVIRT_INSECURE")
	ovirtCafile := os.Getenv("OVIRT_CAFILE")
	ovirtCaBundle := os.Getenv("OVIRT_CA_BUNDLE")
	if ovirtInsecure == "" && ovirtCafile == "" && ovirtCaBundle == "" {
		t.Fatal("OVIRT_INSECURE, OVIRT_CAFILE, or OVIRT_CA_BUNDLE must be set for acceptance tests")
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

func TestSemaphoreProvider(t *testing.T) {
	t.Run("single lock", testSemaphoreProviderSingleLock)
}

func testSemaphoreProviderSingleLock(t *testing.T) {
	lockProvider := newSemaphoreProvider()

	lockComplete := false
	failed := false
	start := make(chan struct{})
	done := make(chan struct{})
	go func() {
		<-start
		lockProvider.Lock("test1", 1)
		if !lockComplete {
			failed = true
		}
		lockProvider.Unlock("test1")
		done <- struct{}{}
	}()
	lockProvider.Lock("test1", 1)
	start <- struct{}{}
	time.Sleep(time.Second)
	lockComplete = true
	lockProvider.Unlock("test1")
	<-done
	if failed {
		t.Fatalf("Lock provider doesn't properly lock.")
	}
}