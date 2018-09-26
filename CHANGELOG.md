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

