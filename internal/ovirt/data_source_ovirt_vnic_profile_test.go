package ovirt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestVNICProfileDataSource(t *testing.T) {
	p := newProvider(newTestLogger(t))
	helper := p.getTestHelper()
	vnicProfileID := helper.GetVNICProfileID()

	vnicProfile, err := helper.GetClient().GetVNICProfile(vnicProfileID)
	if err != nil {
		t.Fatalf("Failed to get VNIC profile: %v", err)
	}

	configByName := fmt.Sprintf(
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

output "vnic_profile_network_id" {
	value = data.ovirt_vnic_profile.test.network_id
}`,
		vnicProfile.Name())

	configWithNetworkID := fmt.Sprintf(
		`
provider "ovirt" {
	mock = true
}

data "ovirt_vnic_profile" "test" {
	name       = "%s"
	network_id = "%s"
}

output "vnic_profile_id" {
	value = data.ovirt_vnic_profile.test.id
}`,
		vnicProfile.Name(),
		vnicProfile.NetworkID())

	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: p.getProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: configByName,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							idVal := s.RootModule().Outputs["vnic_profile_id"].Value.(string)
							if idVal != string(vnicProfile.ID()) {
								return fmt.Errorf("expected VNIC profile ID %s, got %s", vnicProfile.ID(), idVal)
							}

							netIDVal := s.RootModule().Outputs["vnic_profile_network_id"].Value.(string)
							if netIDVal != string(vnicProfile.NetworkID()) {
								return fmt.Errorf("expected Network ID %s, got %s", vnicProfile.NetworkID(), netIDVal)
							}
							return nil
						},
					),
				},
				{
					Config: configWithNetworkID,
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							idVal := s.RootModule().Outputs["vnic_profile_id"].Value.(string)
							if idVal != string(vnicProfile.ID()) {
								return fmt.Errorf("expected VNIC profile ID %s, got %s", vnicProfile.ID(), idVal)
							}
							return nil
						},
					),
				},
				{
					Config: `
provider "ovirt" {
	mock = true
}

data "ovirt_vnic_profile" "test" {
	name = "non_existent_profile_name"
}`,
					ExpectError: regexp.MustCompile("No VNIC profile found with name"),
				},
			},
		},
	)
}

