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
		},

		ResourcesMap: map[string]*schema.Resource{
			"ovh_vrack_publiccloud_attachment":       resourceVRackPublicCloudAttachment(),
			"ovh_publiccloud_private_network":        resourcePublicCloudPrivateNetwork(),
			"ovh_publiccloud_private_network_subnet": resourcePublicCloudPrivateNetworkSubnet(),
			"ovh_publiccloud_user":                   resourcePublicCloudUser(),
		},

		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Endpoint:          d.Get("endpoint").(string),
		ApplicationKey:    d.Get("application_key").(string),
		ApplicationSecret: d.Get("application_secret").(string),
		ConsumerKey:       d.Get("consumer_key").(string),
	}

	if err := config.loadAndValidate(); err != nil {
		return nil, err
	}

	return &config, nil
}
