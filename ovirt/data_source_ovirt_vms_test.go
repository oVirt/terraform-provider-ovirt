// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

//TODO fix test
func DisabledTestAccOvirtVMsDataSource_nameRegexFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)

	id := suite.GenerateRandomID(5)
	diskName := fmt.Sprintf("tf-test-%s", id)
	vmName := fmt.Sprintf("tf-test-%s", id)

	setupContext, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	fh, err := os.Open(suite.TestImageSourcePath())
	if err != nil {
		t.Fatal(fmt.Errorf("failed to open test image file (%w)", err))
	}
	defer func() {
		_ = fh.Close()
	}()
	stat, err := fh.Stat()
	if err != nil {
		t.Fatal(fmt.Errorf("failed to stat test image file (%w)", err))
	}
	result, err := suite.Client().UploadImage(
		setupContext,
		diskName,
		suite.StorageDomainID(),
		true,
		uint64(stat.Size()),
		fh,
	)
	if err != nil {
		t.Fatal(fmt.Errorf("uploading test image failed (%w)", err))
	}
	defer func() {
		if err := suite.Client().RemoveDisk(result.Disk().ID()); err != nil {
			t.Fatal(fmt.Errorf("failed to remove disk image (%w)", err))
		}
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "ovirt_image_transfer" "disk" {
  alias = "%s"
  source_url = "%s"
  storage_domain_id = "%s"
  sparse = true
}

resource "ovirt_vm" "vm" {
  name        = "%s"
  cluster_id  = "%s"
  template_id = "%s"
  auto_start  = false

  os {
    type = "other"
  }

  block_device {
    interface = "virtio"
    disk_id   = ovirt_image_transfer.disk.disk_id
    size      = 1
  }
}

data "ovirt_vms" "name_regex_filtered_vm" {
  name_regex = "^%s$"

  depends_on = [ovirt_vm.vm]
}
`,
					diskName,
					suite.TestImageSourceURL(),
					suite.StorageDomainID(),
					vmName,
					suite.ClusterID(),
					suite.BlankTemplateID(),
					regexp.QuoteMeta(vmName),
				),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_vms.name_regex_filtered_vm"),
					resource.TestCheckResourceAttr("data.ovirt_vms.name_regex_filtered_vm", "vms.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_vms.name_regex_filtered_vm", "vms.0.name", vmName),
				),
			},
		},
	})
}

//TODO fix this test
func DisabledTestAccOvirtVMsDataSource_searchFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)

	id := suite.GenerateRandomID(5)
	diskName := fmt.Sprintf("tf_test_%s", id)
	vmName := fmt.Sprintf("tf_test_%s", id)

	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "ovirt_image_transfer" "disk" {
  alias = "%s"
  source_url = "%s"
  storage_domain_id = "%s"
  sparse = true
}

resource "ovirt_vm" "vm" {
  name        = "%s"
  cluster_id  = "%s"
  template_id = "%s"
  auto_start  = false

  os {
    type = "other"
  }

  block_device {
    interface = "virtio"
    disk_id   = ovirt_image_transfer.disk.disk_id
    size      = 1
  }
}

data "ovirt_vms" "search_filtered_vm" {
  search = {
    criteria       = "name = %s and status = down"
    max            = 1
    case_sensitive = false
  }

  depends_on = [ovirt_vm.vm]
}
`,
					diskName,
					suite.TestImageSourceURL(),
					suite.StorageDomainID(),
					vmName,
					suite.ClusterID(),
					suite.BlankTemplateID(),
					regexp.QuoteMeta(vmName),
				),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_vms.search_filtered_vm"),
					resource.TestCheckResourceAttr("data.ovirt_vms.search_filtered_vm", "vms.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_vms.search_filtered_vm", "vms.0.name", vmName),
				),
			},
		},
	})
}
