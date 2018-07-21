Terraform oVirt Provider plugin
===============================
This plugin allows Terraform to work with the oVirt Virtual Machine management platform.


Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.11.x
-	[Go](https://golang.org/doc/install) 1.9 (to build the provider plugin)


Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/EMSL-MSC/terraform-provider-ovirt`

```sh
$ mkdir -p $GOPATH/src/github.com/EMSL-MSC
$ cd $GOPATH/src/github.com/EMSL-MSC
$ git clone git@github.com:sinokylin/terraform-provider-ovirt
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/EMSL-MSC/terraform-provider-ovirt
$ make build
```


Using the provider
------------------
If you're building the provider, follow the instructions to [install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin) After placing it into your plugins directory,  run `terraform init` to initialize it.

Provider Usage
--------------

* Provider Configuration
```
provider "ovirt" {
  username = "username@profile"
  url = "https://ovirt/ovirt-engine/api"
  password = "Password"
}
```
  * username - (Required) The username to access the oVirt api including the profile used
  * url - (Required) The url to the api endpoint (usually the ovirt server with a path of /ovirt-engine/api)
  * password - (Required) Password to access the server
* Resources
  * ovirt_vm
  * ovirt_disk
  * ovirt_disk_attachment
  * ovirt_datacenter
  * ovirt_network
* Data Sources
  * ovirt_disks
  * ovirt_datacenters
  * ovirt_networks
  * ovirt_clusters


Disclaimer
---------
This material was prepared as an account of work sponsored by an agency of the United States Government.  Neither the United States Government nor the United States Department of Energy, nor Battelle, nor any of their employees, nor any jurisdiction or organization that has cooperated in the development of these materials, makes any warranty, express or implied, or assumes any legal liability or responsibility for the accuracy, completeness, or usefulness or any information, apparatus, product, software, or process disclosed, or represents that its use would not infringe privately owned rights.

Reference herein to any specific commercial product, process, or service by trade name, trademark, manufacturer, or otherwise does not necessarily constitute or imply its endorsement, recommendation, or favoring by the United States Government or any agency thereof, or Battelle Memorial Institute. The views and opinions of authors expressed herein do not necessarily state or reflect those of the United States Government or any agency thereof.

PACIFIC NORTHWEST NATIONAL LABORATORY
operated by
BATTELLE
for the
UNITED STATES DEPARTMENT OF ENERGY
under Contract DE-AC05-76RL01830
