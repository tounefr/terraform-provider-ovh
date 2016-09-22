package ovh

import (
	"fmt"
	"github.com/ovh/go-ovh/ovh"
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

	// targetClient, err := clientWithCK(c, clientDefault)
	// if err != nil {
	// 	return fmt.Errorf("Error getting ovh client with CK: %q\n", err)
	// }

	var me PartialMe
	err = targetClient.Get("/me", &me)
	if err != nil {
		return fmt.Errorf("OVH client seems to be misconfigured: %q\n", err)
	}

	log.Printf("Logged in on OVH API as %s!", me.Firstname)
	c.OVHClient = targetClient

	return nil
}
