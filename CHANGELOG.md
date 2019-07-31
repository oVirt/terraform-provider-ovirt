## 0.4.1 (Jul 31, 2019)

BUG FIXES:

* resource/ovirt_vm: Do not try to start a VM after updating attributes ([#167](https://github.com/oVirt/terraform-provider-ovirt/pull/167))
* resource/ovirt_disk_attachment: Fix failed to check if a disk attachment exists ([#162](https://github.com/oVirt/terraform-provider-ovirt/pull/162))

FEATURES:

* **New Resource:** `ovirt_snapshot` ([#157](https://github.com/oVirt/terraform-provider-ovirt/pull/157))

IMPROVEMENTS:

* doc: Format inline HCL codes in docs ([#164](https://github.com/oVirt/terraform-provider-ovirt/pull/164))
* provider: Add more general method for parsing composite resource ID ([#163](https://github.com/oVirt/terraform-provider-ovirt/pull/163))
* provider: Format the HCL codes definied in acceptance tests ([#160](https://github.com/oVirt/terraform-provider-ovirt/pull/160))


## 0.4.0 (Jul 8, 2019)

BACKWARDS INCOMPATIBILITIES / NOTES:

* provider: This is the first release since it has been transferred to oVirt community under incubation. Please access to the provider with the new ([oVirt/terraform-provider-ovirt](https://github.com/oVirt/terraform-provider-ovirt)).

IMPROVEMENTS:

* provider: Update to Terraform v0.12.2 ([#145](https://github.com/oVirt/terraform-provider-ovirt/pull/145))
* provider: Remove serveral unnecessary scripts in CI process ([#153](https://github.com/oVirt/terraform-provider-ovirt/pull/153))
* provider: Set `GOFLAGS` in CI environment to force `go mod` to use packages under vendor directory ([#155](https://github.com/oVirt/terraform-provider-ovirt/pull/155))


## 0.3.1 (Jun 10, 2019)

BUG FIXES:

* resource/ovirt_vm: Prevent reading VM failure in case of the `original_template` attribute is unavaliable ([#140](https://github.com/imjoey/terraform-provider-ovirt/pull/140))

FEATURES:

* **New Data Source:** `ovirt_hosts` ([#138](https://github.com/imjoey/terraform-provider-ovirt/pull/138))

IMPROVEMENTS:

* provider: Update to Terraform v0.12.1 ([#141](https://github.com/imjoey/terraform-provider-ovirt/pull/141))


## 0.3.0 (May 29, 2019)

BACKWARDS INCOMPATIBILITIES / NOTES:

* provider: This release contains only a Terraform SDK upgrade for compatibility with Terraform v0.12. The provider should remains backwards compatible with Terraform v0.11. This update should have no significant changes in behavior for the provider. Please report any unexpected behavior in new GitHub issues (Terraform oVirt Provider: https://github.com/imjoey/terraform-provider-ovirt/issues) ([#133](https://github.com/imjoey/terraform-provider-ovirt/pull/133))


## 0.2.2 (May 27, 2019)

BUG FIXES:

* resource/ovirt_vm: Prevent creating VM failure and mistaken state diffs due to `memory` attribute


## 0.2.1 (May 22, 2019)

FEATURES:

* **New Resource:** `ovirt_storage_domain` ([#92](https://github.com/imjoey/terraform-provider-ovirt/pull/92))
* **New Resource:** `ovirt_user` ([#98](https://github.com/imjoey/terraform-provider-ovirt/pull/98))
* **New Resource:** `ovirt_cluster` ([#103](https://github.com/imjoey/terraform-provider-ovirt/pull/103))
* **New Resource:** `ovirt_mac_pool` ([#107](https://github.com/imjoey/terraform-provider-ovirt/pull/107))
* **New Resource:** `ovirt_tag` ([#107](https://github.com/imjoey/terraform-provider-ovirt/pull/114))
* **New Resource:** `ovirt_host` ([#121](https://github.com/imjoey/terraform-provider-ovirt/pull/121))
* **New Data Source:** `ovirt_authzs` ([#97](https://github.com/imjoey/terraform-provider-ovirt/pull/97))
* **New Data Source:** `ovirt_users` ([#102](https://github.com/imjoey/terraform-provider-ovirt/pull/102))
* **New Data Source:** `ovirt_mac_pools` ([#109](https://github.com/imjoey/terraform-provider-ovirt/pull/109))
* **New Data Source:** `ovirt_vms` ([#118](https://github.com/imjoey/terraform-provider-ovirt/pull/118))


IMPROVEMENTS:

* provider: Add `header` params support for connection settings ([#72](https://github.com/imjoey/terraform-provider-ovirt/pull/72))
* resource/ovirt_disk: Add `quota_id` attribute support ([#80](https://github.com/imjoey/terraform-provider-ovirt/pull/80))
* doc: Add webswebsite infrastructure and provider documantations ([#81](https://github.com/imjoey/terraform-provider-ovirt/pull/81))
* resource/ovirt_vnic: Add acceptance tests ([#90](https://github.com/imjoey/terraform-provider-ovirt/pull/90))
* resource/ovirt_network: Add acceptance tests ([#91](https://github.com/imjoey/terraform-provider-ovirt/pull/91))
* resource/ovirt_vm: Add `clone` support ([#131](https://github.com/imjoey/terraform-provider-ovirt/pull/131))


## 0.2.0 (September 26, 2018)

BACKWARDS INCOMPATIBILITIES / NOTES:

* provider: All the new or existing resources and data sources have been refactored with the [oVirt Go SDK](https://github.com/imjoey/go-ovirt) to access the oVirt engine API


FEATURES:

* **New Resource:** `ovirt_disk_attachment` ([#1](https://github.com/imjoey/terraform-provider-ovirt/pull/1))
* **New Resource:** `ovirt_datacenter` ([#3](https://github.com/imjoey/terraform-provider-ovirt/pull/3))
* **New Resource:** `ovirt_network` ([#6](https://github.com/imjoey/terraform-provider-ovirt/pull/6))
* **New Resource:** `ovirt_vnic_profile` ([#41](https://github.com/imjoey/terraform-provider-ovirt/pull/41))
* **New Resource:** `ovirt_vnic` ([#56](https://github.com/imjoey/terraform-provider-ovirt/pull/56))
* **New Data Source:** `ovirt_datacenters` ([#4](https://github.com/imjoey/terraform-provider-ovirt/pull/4))
* **New Data Source:** `ovirt_networks` ([#13](https://github.com/imjoey/terraform-provider-ovirt/pull/13))
* **New Data Source:** `ovirt_clusters` ([#26](https://github.com/imjoey/terraform-provider-ovirt/pull/26))
* **New Data Source:** `ovirt_storagedomains` ([#27](https://github.com/imjoey/terraform-provider-ovirt/pull/27))
* **New Data Source:** `ovirt_vnic_profiles` ([#51](https://github.com/imjoey/terraform-provider-ovirt/pull/51))

IMPROVEMENTS:

* provider: Add GNU make integration: ([#15](https://github.com/imjoey/terraform-provider-ovirt/pull/15))
* provider: Add acceptance tests for provider ([#8](https://github.com/imjoey/terraform-provider-ovirt/pull/8))
* provider: Add acceptance tests for all the resources and data sources
* provider: Add travis CI support ([#47](https://github.com/imjoey/terraform-provider-ovirt/pull/47))
* provider: Add missing attributes and processing logic for the existing `ovirt_vm`, `ovirt_disk` resources and `ovirt_disk` data source defined in v0.1.0


## 0.1.0 (March 13, 2018)

BACKWARDS INCOMPATIBILITIES / NOTES:

* Release by [EMSL-MSC](https://github.com/EMSL-MSC/terraform-provider-ovirt/commits) Orgnization, please see [here](https://github.com/EMSL-MSC/terraform-provider-ovirt/releases/tag/0.1.0) for details.


FEATURES:

* **New Resource:** `ovirt_vm`
* **New Resource:** `ovirt_disk`
* **New Data Source:** `ovirt_disk`

