package ovh

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a schema.Provider for OVH.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVH_ENDPOINT", nil),
			},
			"application_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVH_APPLICATION_KEY", ""),
			},
			"application_secret": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVH_APPLICATION_SECRET", ""),
			},
			"consumer_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVH_CONSUMER_KEY", ""),
			},
			"os_auth_url": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_AUTH_URL", nil),
			},
			"os_user_name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_USERNAME", ""),
			},
			"os_tenant_name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_TENANT_NAME", nil),
			},
			"os_password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_PASSWORD", ""),
			},
			"os_endpoint_type": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_ENDPOINT_TYPE", ""),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"ovh_vrack_publiccloud_attachment":       resourceVRackPublicCloudAttachment(),
			"ovh_publiccloud_private_network":        resourcePublicCloudPrivateNetwork(),
			"ovh_publiccloud_private_network_subnet": resourcePublicCloudPrivateNetworkSubnet(),
			"ovh_publiccloud_user":                   resourcePublicCloudUser(),
			"ovh_domain_record":			  resourceDomainRecord(),
		},

		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Endpoint:           d.Get("endpoint").(string),
		ApplicationKey:     d.Get("application_key").(string),
		ApplicationSecret:  d.Get("application_secret").(string),
		ConsumerKey:        d.Get("consumer_key").(string),
		OSIdentityEndpoint: d.Get("os_auth_url").(string),
		OSUsername:         d.Get("os_user_name").(string),
		OSPassword:         d.Get("os_password").(string),
		OSTenantName:       d.Get("os_tenant_name").(string),
		OSEndpointType:     d.Get("os_endpoint_type").(string),
	}

	if err := config.loadAndValidate(); err != nil {
		return nil, err
	}

	return &config, nil
}
