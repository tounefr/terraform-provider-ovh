package ovh

import (
	"fmt"
	"github.com/ovh/go-ovh/ovh"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"log"
)

// Endpoints
const (
	OvhEU = "https://eu.api.ovh.com/1.0"
	OvhCA = "https://ca.api.ovh.com/1.0"
)

var OVHEndpoints = map[string]string{
	"ovh-eu": OvhEU,
	"ovh-ca": OvhCA,
}

type Config struct {
	Endpoint          string
	ApplicationKey    string
	ApplicationSecret string
	ConsumerKey       string
	OVHClient         *ovh.Client

	OSUsername         string
	OSPassword         string
	OSIdentityEndpoint string
	OSTenantName       string
	OSEndpointType     string

	OSClient *gophercloud.ProviderClient
}

/* type used to verify client access to ovh api
 */
type PartialMe struct {
	Firstname string `json:"firstname"`
}

func clientDefault(c *Config) (*ovh.Client, error) {
	if c.ApplicationKey != "" && c.ApplicationSecret != "" {
		client, err := ovh.NewClient(c.Endpoint, c.ApplicationKey, c.ApplicationSecret, c.ConsumerKey)
		if err != nil {
			return nil, err
		}
		return client, nil
	} else {
		client, err := ovh.NewEndpointClient(c.Endpoint)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
}

func (c *Config) loadAndValidate() error {
	validEndpoint := false

	for k, _ := range OVHEndpoints {
		if c.Endpoint == k {
			validEndpoint = true
		}
	}

	if !validEndpoint {
		return fmt.Errorf("%s is not a valid ovh endpoint\n", c.Endpoint)
	}

	targetClient, err := clientDefault(c)
	if err != nil {
		return fmt.Errorf("Error getting ovh client: %q\n", err)
	}

	var me PartialMe
	err = targetClient.Get("/me", &me)
	if err != nil {
		return fmt.Errorf("OVH client seems to be misconfigured: %q\n", err)
	}

	log.Printf("[DEBUG] Logged in on OVH API as %s!", me.Firstname)
	c.OVHClient = targetClient

	log.Printf("[DEBUG] Configuring openstack client!")

	if c.OSEndpointType != "internal" && c.OSEndpointType != "internalURL" &&
		c.OSEndpointType != "admin" && c.OSEndpointType != "adminURL" &&
		c.OSEndpointType != "public" && c.OSEndpointType != "publicURL" &&
		c.OSEndpointType != "" {
		return fmt.Errorf("Invalid openstack endpoint type provided")
	}

	ao := gophercloud.AuthOptions{
		Username:         c.OSUsername,
		Password:         c.OSPassword,
		IdentityEndpoint: c.OSIdentityEndpoint,
		TenantName:       c.OSTenantName,
	}

	client, err := openstack.NewClient(ao.IdentityEndpoint)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Authenticate openstack client with options: %v", ao)
	err = openstack.Authenticate(client, ao)
	if err != nil {
		return err
	}

	c.OSClient = client

	return nil
}

func (c *Config) blockStorageV1Client(region string) (*gophercloud.ServiceClient, error) {
	return openstack.NewBlockStorageV1(c.OSClient, gophercloud.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) blockStorageV2Client(region string) (*gophercloud.ServiceClient, error) {
	return openstack.NewBlockStorageV2(c.OSClient, gophercloud.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) computeV2Client(region string) (*gophercloud.ServiceClient, error) {
	return openstack.NewComputeV2(c.OSClient, gophercloud.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) imageV2Client(region string) (*gophercloud.ServiceClient, error) {
	return openstack.NewImageServiceV2(c.OSClient, gophercloud.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) networkingV2Client(region string) (*gophercloud.ServiceClient, error) {
	return openstack.NewNetworkV2(c.OSClient, gophercloud.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) objectStorageV1Client(region string) (*gophercloud.ServiceClient, error) {
	return openstack.NewObjectStorageV1(c.OSClient, gophercloud.EndpointOpts{
		Region:       region,
		Availability: c.getEndpointType(),
	})
}

func (c *Config) getEndpointType() gophercloud.Availability {
	if c.OSEndpointType == "internal" || c.OSEndpointType == "internalURL" {
		return gophercloud.AvailabilityInternal
	}
	if c.OSEndpointType == "admin" || c.OSEndpointType == "adminURL" {
		return gophercloud.AvailabilityAdmin
	}
	return gophercloud.AvailabilityPublic
}
