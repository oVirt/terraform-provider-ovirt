// Copyright (C) 2018 Joey Ma <majunjiev@gmail.com>
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func resourceOvirtHost() *schema.Resource {
	return &schema.Resource{
		Create: resourceOvirtHostCreate,
		Read:   resourceOvirtHostRead,
		Update: resourceOvirtHostUpdate,
		Delete: resourceOvirtHostDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  "",
			},
			"address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"root_password": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     false,
				ValidateFunc: validation.NoZeroValues,
			},
			"ssh_key_auth": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
				ForceNew: false,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceOvirtHostCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)

	rootPassword, rootPasswordOk := d.GetOk("root_password")
	sshKeyAuth := d.Get("ssh_key_auth").(bool)

	if rootPasswordOk == sshKeyAuth {
		return fmt.Errorf("only one of root_password and ssh_key_auth must be specified")
	}

	hostBuilder := ovirtsdk4.NewHostBuilder().
		Name(d.Get("name").(string)).
		Address(d.Get("address").(string)).
		Description(d.Get("description").(string))

	if rootPasswordOk {
		hostBuilder.RootPassword(rootPassword.(string))
	}

	if sshKeyAuth {
		sshBuilder := ovirtsdk4.NewSshBuilder().
			AuthenticationMethod(ovirtsdk4.SSHAUTHENTICATIONMETHOD_PUBLICKEY)
		hostBuilder.SshBuilder(sshBuilder)
	}

	if clusterID, ok := d.GetOk("cluster_id"); ok {
		hostBuilder.Cluster(ovirtsdk4.NewClusterBuilder().
			Id(clusterID.(string)).
			MustBuild())
	}
	hostsService := conn.SystemService().HostsService()
	addResp, err := hostsService.
		Add().
		Host(hostBuilder.MustBuild()).
		Send()
	if err != nil {
		log.Printf("[DEBUG] Error adding Host (%s): %s", d.Get("name").(string), err)
		return err
	}
	d.SetId(addResp.MustHost().MustId())

	// // Now to activate it
	// _, err = hostsService.HostService(d.Id()).Activate().Send()
	// if err != nil {
	// 	log.Printf("[DEBUG] Error activating Host (%s): %s", d.Id(), err)
	// 	return err
	// }

	log.Printf("[DEBUG] Wait for Host (%s) status to become up", d.Id())
	upStateConf := &resource.StateChangeConf{
		Target:     []string{string(ovirtsdk4.HOSTSTATUS_UP)},
		Refresh:    HostStateRefreshFunc(conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = upStateConf.WaitForState()
	if err != nil {
		log.Printf("[DEBUG] Failed to wait for Host (%s) to become up: %s", d.Id(), err)
		return fmt.Errorf("Error waiting for Host (%s) to be up: %s", d.Id(), err)
	}

	return resourceOvirtHostRead(d, meta)
}

func resourceOvirtHostRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)
	getResp, err := conn.SystemService().
		HostsService().
		HostService(d.Id()).
		Get().
		Send()
	if err != nil {
		if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
			d.SetId("")
			return nil
		}
		return err
	}
	host := getResp.MustHost()
	d.Set("name", host.MustName())
	d.Set("address", host.MustAddress())
	// rootPassword not returned
	if cluster, ok := host.Cluster(); ok {
		d.Set("cluster_id", cluster.MustId())
	}
	if desc, ok := host.Description(); ok {
		d.Set("description", desc)
	}
	if ssh, ok := host.Ssh(); ok {
		if auth, ok := ssh.AuthenticationMethod(); ok {
			if auth == ovirtsdk4.SSHAUTHENTICATIONMETHOD_PUBLICKEY {
				d.Set("ssh_key_auth", true)
			} else {
				d.Set("ssh_key_auth", false)
			}
		}
	}

	return nil
}

func resourceOvirtHostUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)
	attributeUpdate := false
	paramBuilder := ovirtsdk4.NewHostBuilder()

	d.Partial(true)

	if d.HasChange("name") {
		paramBuilder.Name(d.Get("name").(string))
		attributeUpdate = true
	}

	if d.HasChange("description") {
		paramBuilder.Description(d.Get("description").(string))
		attributeUpdate = true
	}

	if d.HasChange("root_password") {
		paramBuilder.RootPassword(d.Get("root_password").(string))
		attributeUpdate = true
	}

	if d.HasChange("ssh_key_auth") {
		method := ovirtsdk4.SSHAUTHENTICATIONMETHOD_PASSWORD
		if d.Get("ssh_key_auth").(bool) {
			method = ovirtsdk4.SSHAUTHENTICATIONMETHOD_PUBLICKEY
		}
		sshBuilder := ovirtsdk4.NewSshBuilder().
			AuthenticationMethod(method)
		paramBuilder.SshBuilder(sshBuilder)
	}

	if attributeUpdate {
		_, err := conn.SystemService().
			HostsService().
			HostService(d.Id()).
			Update().
			Host(paramBuilder.MustBuild()).
			Send()
		if err != nil {
			log.Printf("[DEBUG] Error updating Host (%s): %s", d.Id(), err)
			return err
		}
	}

	d.Partial(false)

	return resourceOvirtHostRead(d, meta)
}

func resourceOvirtHostDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*ovirtsdk4.Connection)
	hostService := conn.SystemService().HostsService().HostService(d.Id())
	getResp, err := hostService.Get().Send()
	if err != nil {
		if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
			return nil
		}
		log.Printf("[DEBUG] Error getting Host (%s): %s", d.Id(), err)
		return err
	}

	currentStatus := getResp.MustHost().MustStatus()
	if currentStatus != ovirtsdk4.HOSTSTATUS_MAINTENANCE {
		log.Printf("[DEBUG] Host (%s) status is %s and now deactivate", d.Id(), currentStatus)
		_, err = hostService.Deactivate().Send()
		if err != nil {
			log.Printf("[DEBUG] Error deactivating Host (%s): %s", d.Id(), err)
			return err
		}
		log.Printf("[DEBUG] Wait for Host (%s) status to become maintenance", d.Id())
		mtStateConf := &resource.StateChangeConf{
			Target:     []string{string(ovirtsdk4.HOSTSTATUS_MAINTENANCE)},
			Refresh:    HostStateRefreshFunc(conn, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutDelete),
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}
		_, err = mtStateConf.WaitForState()
		if err != nil {
			log.Printf("[DEBUG] Failed to wait for Host (%s) to become maintenance: %s", d.Id(), err)
			return fmt.Errorf("Error waiting for Host (%s) to be maintenance: %s", d.Id(), err)
		}
	}

	return resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		log.Printf("[DEBUG] Now to remove Host (%s)", d.Id())
		_, err := hostService.Remove().Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				// Wait until NotFoundError raises
				log.Printf("[DEBUG] Host (%s) has been removed", d.Id())
				return nil
			}
			return resource.RetryableError(fmt.Errorf("Error removing Host (%s): %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("Host (%s) is still being removed", d.Id()))
	})
}

// HostStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// an oVirt Host.
func HostStateRefreshFunc(conn *ovirtsdk4.Connection, hostID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		r, err := conn.SystemService().
			HostsService().
			HostService(hostID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				return nil, "", nil
			}
			return nil, "", err
		}

		return r.MustHost(), string(r.MustHost().MustStatus()), nil
	}
}
