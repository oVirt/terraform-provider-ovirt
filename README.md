Terraform oVirt Provider plugin
===============================

[![Build Status](https://travis-ci.org/imjoey/terraform-provider-ovirt.svg?branch=master)](https://travis-ci.org/imjoey/terraform-provider-ovirt)
[![Go Report Card](https://goreportcard.com/badge/github.com/ovirt/terraform-provider-ovirt)](https://goreportcard.com/report/github.com/ovirt/terraform-provider-ovirt)


This plugin allows Terraform to work with the oVirt Virtual Machine management platform.
It requires oVirt 4.x. 


Statements
-----------

Firstly, this project is inspired by [EMSL-MSC](http://github.com/EMSL-MSC/terraform-provider-ovirt), the author [@Maigard](https://github.com/EMSL-MSC/terraform-provider-ovirt/commits?author=Maigard) surely done a outstanding work and great thanks to him.

While in the last five months, the upstream project was not actively maintained and the pull request I committed is still not reviewed. Since this project is a heavy work in progress, for intuitive and convenient usage, I replaced the references of `EMSL-MSC` with `imjoey` in `main.go`, `README` and some other CI configuration files.

If possible, I would surely be happy to contribute back to the upstream again. ^_^ .


Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.11.x
-	[Go](https://golang.org/doc/install) 1.11 (to build the provider plugin)


Building The Provider
---------------------

```sh
$ git clone git@github.com:ovirt/terraform-provider-ovirt
$ cd terraform-provider-ovirt
$ make build
```

Using the provider
------------------
If you're building the provider, follow the instructions to [install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin) After placing it into your plugins directory,  run `terraform init` to initialize it.

Provider Usage
--------------

* Provider Configuration
```HCL
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
  * ovirt_cluster
  * ovirt_datacenter
  * ovirt_disk
  * ovirt_disk_attachment
  * ovirt_host
  * ovirt_mac_pool
  * ovirt_network
  * ovirt_storage_domain
  * ovirt_tag
  * ovirt_user
  * ovirt_vm
  * ovirt_vnic
  * ovirt_vnic_profile
* Data Sources
  * ovirt_authzs
  * ovirt_clusters
  * ovirt_datacenters
  * ovirt_disks
  * ovirt_hosts
  * ovirt_mac_pools
  * ovirt_networks
  * ovirt_storagedomains
  * ovirt_users
  * ovirt_vms
  * ovirt_vnic_profiles

Provider Documents
--------------
Currently the documents for this provider is not hosted by the official site [Terraform Providers](https://www.terraform.io/docs/providers/index.html). Please enter the provider directory and build the website locally.

```sh
$ make website
```

The commands above will start a docker-based web server powered by [Middleman](https://middlemanapp.com/), which hosts the documents in `website` directory. Simply open `http://localhost:4567/docs/providers/ovirt` and enjoy them.
