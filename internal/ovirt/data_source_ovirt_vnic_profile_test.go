package ovirt

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

func TestVNICProfileDataSource(t *testing.T) {
	t.Parallel()
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

type stubVNICProfile struct {
	ovirtclient.VNICProfile
	id        ovirtclient.VNICProfileID
	name      string
	networkID ovirtclient.NetworkID
}

func (s stubVNICProfile) ID() ovirtclient.VNICProfileID    { return s.id }
func (s stubVNICProfile) Name() string                     { return s.name }
func (s stubVNICProfile) NetworkID() ovirtclient.NetworkID { return s.networkID }

func TestSelectVNICProfileAmbiguous(t *testing.T) {
	t.Parallel()

	profiles := []ovirtclient.VNICProfile{
		stubVNICProfile{id: "11111111-1111-1111-1111-111111111111", name: "ovirtmgmt", networkID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"},
		stubVNICProfile{id: "22222222-2222-2222-2222-222222222222", name: "ovirtmgmt", networkID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"},
	}

	_, diags := selectVNICProfile(profiles, "ovirtmgmt", "", false)
	if !diags.HasError() {
		t.Fatal("expected ambiguity error, got nil")
	}
	summary := diags[0].Summary
	if !strings.Contains(summary, "11111111-1111-1111-1111-111111111111") || !strings.Contains(summary, "22222222-2222-2222-2222-222222222222") {
		t.Fatalf("expected conflicting profile IDs in error message, got: %s", summary)
	}

	selected, diags := selectVNICProfile(profiles, "ovirtmgmt", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", true)
	if diags.HasError() {
		t.Fatalf("unexpected error when disambiguating with network_id: %v", diags)
	}
	if selected.ID() != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("expected second profile, got %s", selected.ID())
	}
}
