package ovirt

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclient "github.com/ovirt/go-ovirt-client/v2"
)

func TestVMDisksResize(t *testing.T) {
	t.Parallel()

	testData := []struct {
		name             string
		diskCount        int
		startingDiskSize uint64
		desiredDiskSize  uint64
	}{
		{
			"empty",
			0,
			uint64(1024 * 1024),
			2 * uint64(1024*1024),
		},
		{
			"single",
			1,
			uint64(1024 * 1024),
			2 * uint64(1024*1024),
		},
		{
			"double",
			2,
			uint64(1024 * 1024),
			2 * uint64(1024*1024),
		},
	}

	for _, testCase := range testData {
		t.Run(
			testCase.name, func(t *testing.T) {
				p := newProvider(newTestLogger(t))

				helper := p.getTestHelper()
				client := helper.GetClient().WithContext(context.Background())
				vm, err := client.CreateVM(
					helper.GetClusterID(),
					helper.GetBlankTemplateID(),
					helper.GenerateTestResourceName(t),
					nil,
				)
				if err != nil {
					t.Fatal(err)
				}

				startingDiskSize := testCase.startingDiskSize
				desiredDiskSize := testCase.desiredDiskSize

				disks := make([]ovirtclient.Disk, testCase.diskCount)

				for i := 0; i < testCase.diskCount; i++ {
					disk, err := client.CreateDisk(
						helper.GetStorageDomainID(),
						ovirtclient.ImageFormatRaw,
						startingDiskSize,
						nil,
					)
					if err != nil {
						t.Fatal(err)
					}

					_, err = client.CreateDiskAttachment(vm.ID(), disk.ID(), ovirtclient.DiskInterfaceVirtIO, nil)
					if err != nil {
						t.Fatal(err)
					}

					disks[i] = disk
				}

				config := fmt.Sprintf(
					`
					provider "ovirt" {
						mock = true
					}
					resource "ovirt_vm_disks_resize" "resized" {
						vm_id = "%s"
						size  = %d
					}`,
					vm.ID(),
					desiredDiskSize,
				)

				resource.UnitTest(
					t, resource.TestCase{
						ProviderFactories: p.getProviderFactories(),
						Steps: []resource.TestStep{
							{
								Config: config,
								Check: func(state *terraform.State) error {
									for _, disk := range disks {
										disk, err := client.GetDisk(disk.ID())
										if err != nil {
											return err
										}
										if disk.ProvisionedSize() != desiredDiskSize {
											return fmt.Errorf(
												"incorrect disk size after apply: %d",
												disk.ProvisionedSize(),
											)
										}
									}
									return nil
								},
							},
							{
								Config:  config,
								Destroy: true,
							},
						},
					},
				)
			},
		)
	}
}
