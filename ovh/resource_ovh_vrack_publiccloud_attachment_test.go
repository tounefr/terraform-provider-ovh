package ovh

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"testing"
)

var testAccVRackPublicCloudAttachmentConfig = fmt.Sprintf(`
resource "ovh_vrack_public_cloud_attachment" "attach" {
  vrack_id = "%s"
	project_id = "%s"
}
`, os.Getenv("OVH_VRACK"), os.Getenv("OVH_PUBLIC_CLOUD"))

func TestAccVRackPublicCloudAttachment_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccCheckVRackPublicCloudAttachmentPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVRackPublicCloudAttachmentDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccVRackPublicCloudAttachmentConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVRackPublicCloudAttachmentExists("ovh_vrack_public_cloud_attachment.attach", t),
				),
			},
		},
	})
}

func testAccCheckVRackPublicCloudAttachmentPreCheck(t *testing.T) {
	testAccPreCheck(t)
	testAccCheckVRackExists(t)
	testAccCheckPublicCloudExists(t)
}

func testAccCheckVRackExists(t *testing.T) {
	type vrackResponse struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	r := vrackResponse{}

	endpoint := fmt.Sprintf("/vrack/%s", os.Getenv("OVH_VRACK"))

	err := testAccOVHClient.Get(endpoint, &r)
	if err != nil {
		t.Fatalf("Error: %q\n", err)
	}
	t.Logf("Read VRack %s -> name:'%s', desc:'%s' ", endpoint, r.Name, r.Description)

}

func testAccCheckPublicCloudExists(t *testing.T) {
	type cloudProjectResponse struct {
		ID          string `json:"project_id"`
		Status      string `json:"status"`
		Description string `json:"description"`
	}

	r := cloudProjectResponse{}

	endpoint := fmt.Sprintf("/cloud/project/%s", os.Getenv("OVH_PUBLIC_CLOUD"))

	err := testAccOVHClient.Get(endpoint, &r)
	if err != nil {
		t.Fatalf("Error: %q\n", err)
	}
	t.Logf("Read Cloud Project %s -> status: '%s', desc: '%s'", endpoint, r.Status, r.Description)

}

func testAccCheckVRackPublicCloudAttachmentExists(n string, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.Attributes["vrack_id"] == "" {
			return fmt.Errorf("No VRack ID is set")
		}

		if rs.Primary.Attributes["project_id"] == "" {
			return fmt.Errorf("No Project ID is set")
		}

		return vrackPublicCloudAttachmentExists(rs.Primary.Attributes["vrack_id"], rs.Primary.Attributes["project_id"], config.OVHClient)
	}
}

func testAccCheckVRackPublicCloudAttachmentDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovh_vrack_public_cloud_attachment" {
			continue
		}

		err := vrackPublicCloudAttachmentExists(rs.Primary.Attributes["vrack_id"], rs.Primary.Attributes["project_id"], config.OVHClient)
		if err == nil {
			return fmt.Errorf("VRack > Public Cloud Attachment still exists")
		}

	}
	return nil
}
