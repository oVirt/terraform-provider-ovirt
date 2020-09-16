package ovirt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func TestAccOvirtSnapshot_basic(t *testing.T) {
	description := "description for snapshot"
	saveMemory := true

	var snapshot ovirtsdk4.Snapshot
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "ovirt_snapshot.snapshot",
		CheckDestroy:  testAccCheckSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotBasic(description, saveMemory),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOvirtSnapshotExists("ovirt_snapshot.snapshot", &snapshot),
					resource.TestCheckResourceAttr("ovirt_snapshot.snapshot", "description", description),
					resource.TestCheckResourceAttr("ovirt_snapshot.snapshot", "save_memory", fmt.Sprintf("%t", saveMemory)),
					resource.TestCheckResourceAttr("ovirt_snapshot.snapshot", "status", string(ovirtsdk4.SNAPSHOTSTATUS_OK)),
				),
			},
		},
	})
}

func testAccCheckSnapshotDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovirt_snapshot" {
			continue
		}

		parts, err := parseResourceID(rs.Primary.ID, 2)
		if err != nil {
			return err
		}
		vmID, snapshotID := parts[0], parts[1]

		getResp, err := conn.SystemService().
			VmsService().
			VmService(vmID).
			SnapshotsService().
			SnapshotService(snapshotID).
			Get().
			Send()

		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				continue
			}
			return err
		}
		if _, ok := getResp.Snapshot(); ok {
			return fmt.Errorf("Snapshot %s still exist", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckOvirtSnapshotExists(n string, v *ovirtsdk4.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Snapshot ID is set")
		}

		parts, err := parseResourceID(rs.Primary.ID, 2)
		if err != nil {
			return err
		}
		vmID, snapshotID := parts[0], parts[1]

		conn := testAccProvider.Meta().(*ovirtsdk4.Connection)
		getResp, err := conn.SystemService().
			VmsService().
			VmService(vmID).
			SnapshotsService().
			SnapshotService(snapshotID).
			Get().
			Send()
		if err != nil {
			return err
		}
		snapshot, ok := getResp.Snapshot()
		if ok {
			*v = *snapshot
			return nil
		}
		return fmt.Errorf("Snapshot %s not exist", rs.Primary.ID)
	}
}

func testAccSnapshotBasic(description string, saveMemory bool) string {
	return fmt.Sprintf(`
data "ovirt_vms" "v" {
  search = {
    criteria = "name = HostedEngine"
  }
}

locals {
  vm_id = data.ovirt_vms.v.vms.0.id
}

resource "ovirt_snapshot" "snapshot" {
  description = "%s"
  vm_id       = local.vm_id
  save_memory = %t
}
`, description, saveMemory)
}
