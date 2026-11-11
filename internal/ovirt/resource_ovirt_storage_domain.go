//nolint:revive
package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client/v3"
)

var storageDomainSchema = map[string]*schema.Schema{
	"id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "oVirt ID of this storage domain.",
	},
	"name": {
		Type:             schema.TypeString,
		Required:         true,
		Description:      "User-provided name for the storage domain.",
		ValidateDiagFunc: validateNonEmpty,
	},
	"host_id": {
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		Description:      "Host ID used to deploy and connect the storage domain.",
		ValidateDiagFunc: validateUUID,
	},
	"storage_type": {
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "nfs",
		ForceNew:    true,
		Description: "Storage type (defaults to 'nfs').",
	},
	"storage_address": {
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		Description:      "Storage server address or hostname.",
		ValidateDiagFunc: validateNonEmpty,
	},
	"storage_path": {
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		Description:      "Storage path or export on the server.",
		ValidateDiagFunc: validateNonEmpty,
	},
	"storage_format": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "Storage format version (e.g. 'v5').",
	},
	"description": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "User-provided description for the storage domain.",
	},
	"comment": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "User-provided comment for the storage domain.",
	},
	"wipe_after_delete": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "Wipe after delete by default for disks in this storage domain.",
	},
	"backup": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "Designate storage domain for backup.",
	},
	"critical_space_action_blocker": {
		Type:             schema.TypeInt,
		Optional:         true,
		Computed:         true,
		ValidateDiagFunc: validatePositiveInt,
		Description:      "Percentage threshold for blocking actions on critical space.",
	},
	"warning_low_space_indicator": {
		Type:             schema.TypeInt,
		Optional:         true,
		Computed:         true,
		ValidateDiagFunc: validatePositiveInt,
		Description:      "Percentage threshold for low space warning indicator.",
	},
	"destroy_on_delete": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Force removal of the storage domain even if not accessible.",
	},
	"format_on_delete": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Format the underlying storage when deleting.",
	},
}

func (p *provider) storageDomainResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: p.storageDomainCreate,
		ReadContext:   p.storageDomainRead,
		UpdateContext: p.storageDomainUpdate,
		DeleteContext: p.storageDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema:      storageDomainSchema,
		Description: "The ovirt_storage_domain resource creates and manages storage domains in oVirt.",
	}
}

func (p *provider) storageDomainCreate(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	name := data.Get("name").(string)
	hostID := data.Get("host_id").(string)
	address := data.Get("storage_address").(string)
	path := data.Get("storage_path").(string)

	params := ovirtclient.NewCreateNFSStorageDomainParams()
	if desc, ok := data.GetOk("description"); ok {
		var err error
		params, err = params.WithDescription(desc.(string))
		if err != nil {
			return errorToDiags("set description", err)
		}
	}
	if comment, ok := data.GetOk("comment"); ok {
		var err error
		params, err = params.WithComment(comment.(string))
		if err != nil {
			return errorToDiags("set comment", err)
		}
	}
	if sf, ok := data.GetOk("storage_format"); ok {
		var err error
		params, err = params.WithStorageFormat(ovirtclient.StorageFormat(sf.(string)))
		if err != nil {
			return errorToDiags("set storage_format", err)
		}
	}
	if wad, ok := data.GetOkExists("wipe_after_delete"); ok {
		var err error
		params, err = params.WithWipeAfterDelete(wad.(bool))
		if err != nil {
			return errorToDiags("set wipe_after_delete", err)
		}
	}
	if b, ok := data.GetOkExists("backup"); ok {
		var err error
		params, err = params.WithBackup(b.(bool))
		if err != nil {
			return errorToDiags("set backup", err)
		}
	}
	if csab, ok := data.GetOk("critical_space_action_blocker"); ok {
		var err error
		params, err = params.WithCriticalSpaceActionBlocker(int64(csab.(int)))
		if err != nil {
			return errorToDiags("set critical_space_action_blocker", err)
		}
	}
	if wlsi, ok := data.GetOk("warning_low_space_indicator"); ok {
		var err error
		params, err = params.WithWarningLowSpaceIndicator(int64(wlsi.(int)))
		if err != nil {
			return errorToDiags("set warning_low_space_indicator", err)
		}
	}

	sd, err := client.CreateNFSStorageDomain(name, ovirtclient.HostID(hostID), address, path, params)
	if err != nil {
		return errorToDiags(fmt.Sprintf("create storage domain %s", name), err)
	}

	return storageDomainToData(sd, data)
}

func (p *provider) storageDomainRead(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	sd, err := client.GetStorageDomain(ovirtclient.StorageDomainID(data.Id()))
	if err != nil {
		if isNotFound(err) {
			data.SetId("")
			return nil
		}
		return errorToDiags(fmt.Sprintf("get storage domain %s", data.Id()), err)
	}
	return storageDomainToData(sd, data)
}

func (p *provider) storageDomainUpdate(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)

	params := ovirtclient.NewUpdateStorageDomainParams()
	if data.HasChange("name") {
		var err error
		params, err = params.WithName(data.Get("name").(string))
		if err != nil {
			return errorToDiags("set name", err)
		}
	}
	if data.HasChange("description") {
		var err error
		params, err = params.WithDescription(data.Get("description").(string))
		if err != nil {
			return errorToDiags("set description", err)
		}
	}
	if data.HasChange("comment") {
		var err error
		params, err = params.WithComment(data.Get("comment").(string))
		if err != nil {
			return errorToDiags("set comment", err)
		}
	}
	if data.HasChange("wipe_after_delete") {
		if wad, ok := data.GetOkExists("wipe_after_delete"); ok {
			var err error
			params, err = params.WithWipeAfterDelete(wad.(bool))
			if err != nil {
				return errorToDiags("set wipe_after_delete", err)
			}
		}
	}
	if data.HasChange("backup") {
		if b, ok := data.GetOkExists("backup"); ok {
			var err error
			params, err = params.WithBackup(b.(bool))
			if err != nil {
				return errorToDiags("set backup", err)
			}
		}
	}
	if data.HasChange("critical_space_action_blocker") {
		if csab, ok := data.GetOk("critical_space_action_blocker"); ok {
			var err error
			params, err = params.WithCriticalSpaceActionBlocker(int64(csab.(int)))
			if err != nil {
				return errorToDiags("set critical_space_action_blocker", err)
			}
		}
	}
	if data.HasChange("warning_low_space_indicator") {
		if wlsi, ok := data.GetOk("warning_low_space_indicator"); ok {
			var err error
			params, err = params.WithWarningLowSpaceIndicator(int64(wlsi.(int)))
			if err != nil {
				return errorToDiags("set warning_low_space_indicator", err)
			}
		}
	}

	sd, err := client.UpdateStorageDomain(ovirtclient.StorageDomainID(data.Id()), params)
	if err != nil {
		return errorToDiags(fmt.Sprintf("update storage domain %s", data.Id()), err)
	}
	return storageDomainToData(sd, data)
}

func (p *provider) storageDomainDelete(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	client := p.client.WithContext(ctx)
	destroy := data.Get("destroy_on_delete").(bool)
	format := data.Get("format_on_delete").(bool)

	params := ovirtclient.NewRemoveStorageDomainParams().
		MustWithDestroy(destroy).
		MustWithFormat(format)

	if err := client.RemoveStorageDomain(ovirtclient.StorageDomainID(data.Id()), params); err != nil {
		if isNotFound(err) {
			data.SetId("")
			return nil
		}
		return errorToDiags(fmt.Sprintf("remove storage domain %s", data.Id()), err)
	}
	data.SetId("")
	return nil
}

func storageDomainToData(sd ovirtclient.StorageDomain, data *schema.ResourceData) diag.Diagnostics {
	diags := diag.Diagnostics{}
	data.SetId(string(sd.ID()))
	diags = appendDiags(diags, "set name on ovirt_storage_domain", data.Set("name", sd.Name()))
	diags = appendDiags(diags, "set storage_type on ovirt_storage_domain", data.Set("storage_type", string(sd.StorageType())))
	if desc := sd.Description(); desc != nil {
		diags = appendDiags(diags, "set description on ovirt_storage_domain", data.Set("description", *desc))
	} else {
		diags = appendDiags(diags, "clear description on ovirt_storage_domain", data.Set("description", nil))
	}
	if comment := sd.Comment(); comment != nil {
		diags = appendDiags(diags, "set comment on ovirt_storage_domain", data.Set("comment", *comment))
	} else {
		diags = appendDiags(diags, "clear comment on ovirt_storage_domain", data.Set("comment", nil))
	}
	if addr := sd.StorageAddress(); addr != nil {
		diags = appendDiags(diags, "set storage_address on ovirt_storage_domain", data.Set("storage_address", *addr))
	}
	if path := sd.StoragePath(); path != nil {
		diags = appendDiags(diags, "set storage_path on ovirt_storage_domain", data.Set("storage_path", *path))
	}
	if sf := sd.StorageFormat(); sf != nil {
		diags = appendDiags(diags, "set storage_format on ovirt_storage_domain", data.Set("storage_format", string(*sf)))
	}
	if wad := sd.WipeAfterDelete(); wad != nil {
		diags = appendDiags(diags, "set wipe_after_delete on ovirt_storage_domain", data.Set("wipe_after_delete", *wad))
	}
	if b := sd.Backup(); b != nil {
		diags = appendDiags(diags, "set backup on ovirt_storage_domain", data.Set("backup", *b))
	}
	if csab := sd.CriticalSpaceActionBlocker(); csab != nil {
		diags = appendDiags(diags, "set critical_space_action_blocker on ovirt_storage_domain", data.Set("critical_space_action_blocker", int(*csab)))
	}
	if wlsi := sd.WarningLowSpaceIndicator(); wlsi != nil {
		diags = appendDiags(diags, "set warning_low_space_indicator on ovirt_storage_domain", data.Set("warning_low_space_indicator", int(*wlsi)))
	}
	return diags
}
