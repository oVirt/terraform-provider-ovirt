package ovirt_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

//TODO fix test
func DisabledTestAccOvirtTemplatesDataSource_nameRegexFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	id := suite.GenerateRandomID(5)
	diskName := fmt.Sprintf("terraform_test_%s_disk", id)
	templateName := fmt.Sprintf("terraform_test_%s_template", id)
	tplVmName := fmt.Sprintf("terraform_test_%s_template_vm", id)
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: suite.TerraformFromTemplate(`
{{- $suite := index $ "suite" -}}
{{- $diskName := index $ "diskName" -}}
{{- $tplName := index $ "tplName" -}}
{{- $tplVmName := index $ "tplVmName" -}}
resource "ovirt_image_transfer" "disk" {
  alias             = "{{ $diskName }}"
  source_url        = "{{ $suite.TestImageSourceURL }}"
  storage_domain_id = "{{ $suite.StorageDomainID }}"
  sparse            = true
}

resource "ovirt_vm" "vm" {
  name        = "{{ $tplVmName }}"
  cluster_id  = "{{ $suite.ClusterID }}"
  template_id = "{{ $suite.BlankTemplateID }}"
  auto_start  = false

  os {
    type = "other"
  }

  block_device {
	storage_domain = "{{ $suite.StorageDomain.MustName }}"
    interface = "virtio"
    disk_id   = ovirt_image_transfer.disk.disk_id
    size      = 1
  }
}

resource "ovirt_template" "test" {
  name       = "{{ $tplName }}"
  cluster_id = "{{ $suite.ClusterID }}"
  cores      = 1
  threads    = 1
  sockets    = 1
  vm_id      = ovirt_vm.vm.id
}

data "ovirt_templates" "name_regex_filtered_template" {
  name_regex = "^{{ $tplName | quoteRegexp }}$"
  depends_on = [ovirt_template.test]
}
`,
					map[string]interface{}{
						"suite":     suite,
						"tplVmName": tplVmName,
						"tplName":   templateName,
						"diskName":  diskName,
					}),
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_templates.name_regex_filtered_template"),
					resource.TestCheckResourceAttr(
						"data.ovirt_templates.name_regex_filtered_template",
						"templates.#",
						"1",
					),
					resource.TestMatchResourceAttr(
						"data.ovirt_templates.name_regex_filtered_template",
						"templates.0.name",
						regexp.MustCompile(fmt.Sprintf("^%s$", templateName)),
					),
				),
			},
		},
	})
}

//TODO fix test
func DisabledTestAccOvirtTemplatesDataSource_searchFilter(t *testing.T) {
	suite := getOvirtTestSuite(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  suite.PreCheck,
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOvirtTemplatesDataSourceSearchConfig,
				Check: resource.ComposeTestCheckFunc(
					suite.TestDataSource("data.ovirt_templates.search_filtered_template"),
					resource.TestCheckResourceAttr("data.ovirt_templates.search_filtered_template", "templates.#", "1"),
					resource.TestCheckResourceAttr("data.ovirt_templates.search_filtered_template", "templates.0.name", "centOST"),
				),
			},
		},
	})

}

var testAccCheckOvirtTemplatesDataSourceSearchConfig = `
data "ovirt_templates" "search_filtered_template" {
  search = {
    criteria       = "name = centOST"
    max            = 2
    case_sensitive = false
  }
}
`
