package ovirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func TestVnicProfileID(t *testing.T) {
	p := newProvider(newTestLogger(t))
	helper := p.getTestHelper()
	vnicProfileID := helper.GetVNICProfileID()
	// FIXME: How should I handle this error?
	vnicProfile, _ := helper.GetClient().GetVNICProfile(ovirtclient.VNICProfileID(vnicProfileID))

	config := fmt.Sprintf(
		`
provider "ovirt" {
	mock = true
}

data "ovirt_vnic_profile" "test" {
	name = "%s"
}

output "vnic_profile_id" {
	value = data.ovirt_vnic_profile.test.id
}

output "vnic_profile_name" {
	value = data.ovirt_vnic_profile.test.name
}
`, vnicProfile.Name())

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							v := s.RootModule().Outputs["vnic_profile_id"].Value.(string)
							tpl, err := p.getTestHelper().GetClient().GetVNICProfile(ovirtclient.VNICProfileID(v))
							if err != nil {
								return err
							}

							v = tpl.Name()
							if v != string(vnicProfileID) {
								return err
							}
							return nil
						},
					),
				},
			},
		},
	)
}
