package ovh

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/ovh/go-ovh/ovh"
	"log"
	"os"
	"testing"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider
var testAccOVHClient *ovh.Client

func init() {
	log.SetOutput(os.Stdout)
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"ovh": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	v := os.Getenv("OVH_ENDPOINT")
	if v == "" {
		t.Fatal("OVH_ENDPOINT must be set for acceptance tests")
	}

	v = os.Getenv("OVH_APPLICATION_KEY")
	if v == "" {
		t.Fatal("OVH_APPLICATION_KEY must be set for acceptance tests")
	}

	v = os.Getenv("OVH_APPLICATION_SECRET")
	if v == "" {
		t.Fatal("OVH_APPLICATION_SECRET must be set for acceptance tests")
	}

	v = os.Getenv("OVH_CONSUMER_KEY")
	if v == "" {
		t.Fatal("OVH_CONSUMER_KEY must be set for acceptance tests")
	}

	v = os.Getenv("OVH_VRACK")
	if v == "" {
		t.Fatal("OVH_VRACK must be set for acceptance tests")
	}

	v = os.Getenv("OVH_PUBLIC_CLOUD")
	if v == "" {
		t.Fatal("OVH_PUBLIC_CLOUD must be set for acceptance tests")
	}

	if testAccOVHClient == nil {
		config := Config{
			Endpoint:          os.Getenv("OVH_ENDPOINT"),
			ApplicationKey:    os.Getenv("OVH_APPLICATION_KEY"),
			ApplicationSecret: os.Getenv("OVH_APPLICATION_SECRET"),
			ConsumerKey:       os.Getenv("OVH_CONSUMER_KEY"),
		}

		if err := config.loadAndValidate(); err != nil {
			t.Fatalf("couln't load OVH Client: %s", err)
		} else {
			testAccOVHClient = config.OVHClient
		}
	}
}
