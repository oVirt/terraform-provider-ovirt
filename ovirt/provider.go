package ovirt

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v3"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

var providerSchema = map[string]*schema.Schema{
	"username": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Username and realm for oVirt authentication. Required when mock = false. Example: `admin@internal`",
	},
	"password": {
		Type:        schema.TypeString,
		Optional:    true,
		Sensitive:   true,
		Description: "Password for oVirt authentication. Required when mock = false.",
	},
	"url": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "URL for the oVirt engine API. Required when mock = false. Example: `https://example.com/ovirt-engine/api/`",
	},
	"extra_headers": {
		Type:        schema.TypeMap,
		Optional:    true,
		Elem:        schema.TypeString,
		Description: "Additional HTTP headers to set on each API call.",
	},
	"tls_insecure": {
		Type:             schema.TypeBool,
		Optional:         true,
		ValidateDiagFunc: validateTLSInsecure,
		Description:      "Disable certificate verification when connecting the Engine. This is not recommended. Setting this option is incompatible with other `tls_` options.",
	},
	"tls_system": {
		Type:             schema.TypeBool,
		Optional:         true,
		ValidateDiagFunc: validateTLSSystem,
		Description:      "Use the system certificate pool to verify the Engine certificate. This does not work on Windows. Can be used in parallel with other `tls_` options, one tls_ option is required when mock = false.",
	},
	"tls_ca_bundle": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Validate the Engine certificate against the provided CA certificates. The certificate chain passed should be in PEM format. Can be used in parallel with other `tls_` options, one `tls_` option is required when mock = false.",
	},
	"tls_ca_files": {
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		Description: "Validate the Engine certificate against the CA certificates provided in the files in this parameter. The files should contain certificates in PEM format. Can be used in parallel with other tls_ options, one tls_ option is required when mock = false.",
		// Validating TypeList fields is not yet supported in Terraform.
		//ValidateDiagFunc: validateFilesExist,
	},
	"tls_ca_dirs": {
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		Description: "Validate the engine certificate against the CA certificates provided in the specified directories. The directory should contain only files with certificates in PEM format. Can be used in parallel with other tls_ options, one tls_ option is required when mock = false.",
		// Validating TypeList fields is not yet supported in Terraform.
		//ValidateDiagFunc: validateDirsExist,
	},
	"mock": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set to true, the Terraform provider runs against an internal simulation. This should only be used for testing when an oVirt engine is not available as the mock backend does not persist state across runs. When set to false, one of the tls_ options is required.",
	},
}

// New returns a new Terraform provider schema for oVirt.
func New() func() *schema.Provider {
	return newProvider(newTerraformLogger()).getProvider
}

func newProvider(logger ovirtclientlog.Logger) providerInterface {
	helper, err := ovirtclient.NewMockTestHelper(
		logger,
	)
	if err != nil {
		panic(err)
	}
	return &provider{
		testHelper: helper,
	}
}

type providerInterface interface {
	getTestHelper() ovirtclient.TestHelper
	getProvider() *schema.Provider
	getProviderFactories() map[string]func() (*schema.Provider, error)
}

type provider struct {
	testHelper ovirtclient.TestHelper
	client     ovirtclient.Client
}

func (p *provider) getTestHelper() ovirtclient.TestHelper {
	return p.testHelper
}

func (p *provider) getProvider() *schema.Provider {
	return &schema.Provider{
		Schema:               providerSchema,
		ConfigureContextFunc: p.configureProvider,
		ResourcesMap: map[string]*schema.Resource{
			"ovirt_affinity_group":           p.affinityGroupResource(),
			"ovirt_vm":                       p.vmResource(),
			"ovirt_vm_start":                 p.vmStartResource(),
			"ovirt_vm_tag":                   p.vmTagResource(),
			"ovirt_vm_optimize_cpu_settings": p.vmOptimizeCPUSettingsResource(),
			"ovirt_disk":                     p.diskResource(),
			"ovirt_disk_attachment":          p.diskAttachmentResource(),
			"ovirt_disk_attachments":         p.diskAttachmentsResource(),
			"ovirt_nic":                      p.nicResource(),
			"ovirt_tag":                      p.tagResource(),
		},
		DataSourcesMap: map[string]*schema.Resource{},
	}
}

func (p *provider) getProviderFactories() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"ovirt": func() (*schema.Provider, error) { //nolint:unparam
			return p.getProvider(), nil
		},
	}
}

func (p *provider) configureProvider(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if mock, ok := data.GetOk("mock"); ok && mock == true {
		p.client = p.testHelper.GetClient()
		return p, diags
	}

	url, diags := extractString(data, "url", diags)
	username, diags := extractString(data, "username", diags)
	password, diags := extractString(data, "password", diags)

	tls := ovirtclient.TLS()
	if insecure, ok := data.GetOk("tls_insecure"); ok && insecure == true {
		tls.Insecure()
	}
	if system, ok := data.GetOk("tls_system"); ok && system == true {
		tls.CACertsFromSystem()
	}
	if caFiles, ok := data.GetOk("tls_ca_files"); ok {
		caFileList, ok := caFiles.([]string)
		if !ok {
			diags = append(
				diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "The tls_ca_files option is not a list of files",
					Detail:   "The tls_ca_files option must be a list of files containing PEM-formatted certificates",
				},
			)
		} else {
			for _, caFile := range caFileList {
				tls.CACertsFromFile(caFile)
			}
		}
	}
	if caDirs, ok := data.GetOk("tls_ca_dirs"); ok {
		caDirList, ok := caDirs.([]string)
		if !ok {
			diags = append(
				diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "The tls_ca_dirs option is not a list of files",
					Detail:   "The tls_ca_dirs option must be a list of files containing PEM-formatted certificates",
				},
			)
		} else {
			for _, caDir := range caDirList {
				tls.CACertsFromDir(caDir)
			}
		}
	}
	if caBundle, ok := data.GetOk("tls_ca_bundle"); ok {
		caCerts, ok := caBundle.(string)
		if !ok {
			diags = append(
				diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "The tls_ca_bundle option is not a string",
					Detail:   "The tls_ca_bundle option must be a string containing the CA certificates in PEM format",
				},
			)
		} else {
			tls.CACertsFromMemory([]byte(caCerts))
		}
	}

	if len(diags) != 0 {
		return nil, diags
	}

	client, err := ovirtclient.New(
		url,
		username,
		password,
		tls,
		newTerraformLogger().WithContext(ctx),
		nil,
	)
	if err != nil {
		diags = append(
			diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Failed to create oVirt client",
				Detail:        err.Error(),
				AttributePath: nil,
			},
		)
		return nil, diags
	}
	p.client = client
	return p, diags
}
