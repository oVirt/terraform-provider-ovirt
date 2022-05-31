package ovirt

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestVMAffinityGroupResource(t *testing.T) {
	t.Parallel()

	p := newProvider(newTestLogger(t))
	clusterID := p.getTestHelper().GetClusterID()
	affinityGroupName := t.Name() + "_" + p.getTestHelper().GenerateRandomID(5)
	vmName := t.Name() + "_" + p.getTestHelper().GenerateRandomID(5)

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(
						`
						provider "ovirt" {
							mock = true
						}

						resource "ovirt_affinity_group" "ag1" {
							cluster_id = "%s"
							name = "%s"
						}

						resource "ovirt_vm" "vm1" {
							cluster_id  = "%s"
							template_id = "%s"
							name        = "%s"
						}

						resource "ovirt_vm_affinity_group" "vm1_to_ag1" {
							cluster_id = "%s"
							vm_id = ovirt_vm.vm1.id
							affinity_group_id = ovirt_affinity_group.ag1.id

							depends_on = [ovirt_vm.vm1, ovirt_affinity_group.ag1]
						}
						`,
						clusterID,
						affinityGroupName,
						clusterID,
						p.getTestHelper().GetBlankTemplateID(),
						vmName,
						clusterID,
					),
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(
							"ovirt_affinity_group.ag1",
							"name",
							regexp.MustCompile(affinityGroupName),
						),
						func(s *terraform.State) error {
							VMID := s.RootModule().Resources["ovirt_vm.vm1"].Primary.Attributes["id"]

							affinityGroup, err := p.getTestHelper().GetClient().WithContext(context.Background()).GetAffinityGroupByName(clusterID, affinityGroupName)
							if err != nil {
								return fmt.Errorf("Failed to get affinity group '%s' by name", affinityGroupName)
							}
							for _, affinityGroupVMID := range affinityGroup.VMIDs() {
								if string(affinityGroupVMID) == VMID {
									return nil
								}
							}
							return fmt.Errorf("VM '%s' not found in affinity groups '%s' list of VMs '%v'", VMID, affinityGroupName, affinityGroup.VMIDs())
						},
					),
				},
			},
		},
	)
}
