package ovh

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ovh/go-ovh/ovh"
	"log"
	"regexp"
	"time"
)

var vpcaID = regexp.MustCompile("vrack_(.+)-cloudproject_(.+)-attach")

func resourceVRackPublicCloudAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceVRackPublicCloudAttachmentCreate,
		Read:   resourceVRackPublicCloudAttachmentRead,
		Delete: resourceVRackPublicCloudAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				params := vpcaID.FindStringSubmatch(d.Id())
				if params == nil {
					return nil, fmt.Errorf("[ERROR] couln't extract vrack id nor project id from id %q", d.Id())
				}

				d.Set("vrack_id", params[1])
				d.Set("project_id", params[2])

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"vrack_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVH_VRACK_ID", ""),
			},
			"project_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVH_PROJECT_ID", ""),
			},
		},
	}
}

// Params
type attachParams struct {
	Project string `json:"project"`
}

// Task Params
type taskParams struct {
	ServiceName string `json:"serviceName"`
	TaskId      string `json:"taskId"`
}

type attachTaskResponse struct {
	Id           int       `json:"id"`
	Function     string    `json:"function"`
	TargetDomain string    `json:"targetDomain"`
	Status       string    `json:"status"`
	ServiceName  string    `json:"serviceName"`
	OrderId      int       `json:"orderId"`
	LastUpdate   time.Time `json:"lastUpdate"`
	TodoDate     time.Time `json:"TodoDate"`
}

func resourceVRackPublicCloudAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	vrackId := d.Get("vrack_id").(string)
	params := &attachParams{Project: d.Get("project_id").(string)}
	r := attachTaskResponse{}

	log.Printf("[DEBUG] Will Attach VRack %s -> PublicCloud %s", vrackId, params.Project)

	endpoint := fmt.Sprintf("/vrack/%s/cloudProject", vrackId)

	err := config.OVHClient.Post(endpoint, params, &r)
	if err != nil {
		return fmt.Errorf("Error calling %s with params %s:\n\t %q", endpoint, params, err)
	}
	log.Printf("[DEBUG] Waiting for Attachement Task id %d: VRack %s ->  PublicCloud %s", r.Id, vrackId, params.Project)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"init", "todo", "doing"},
		Target:     []string{"completed"},
		Refresh:    VRackTaskRefreshFunc(config.OVHClient, vrackId, r.Id),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for vrack (%s) to attach to public cloud (%s): %s", vrackId, params.Project, err)
	}
	log.Printf("[DEBUG] Created Attachement Task id %d: VRack %s ->  PublicCloud %s", r.Id, vrackId, params.Project)

	//set id
	d.SetId(fmt.Sprintf("vrack_%s-cloudproject_%s-attach", vrackId, params.Project))

	return nil
}

func resourceVRackPublicCloudAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	vrackId := d.Get("vrack_id").(string)
	params := &attachParams{Project: d.Get("project_id").(string)}
	r := attachTaskResponse{}
	endpoint := fmt.Sprintf("/vrack/%s/cloudProject/%s", vrackId, params.Project)

	err := config.OVHClient.Get(endpoint, &r)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Read VRack %s ->  PublicCloud %s", vrackId, params.Project)

	return nil
}

func resourceVRackPublicCloudAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	vrackId := d.Get("vrack_id").(string)
	params := &attachParams{Project: d.Get("project_id").(string)}

	r := attachTaskResponse{}
	endpoint := fmt.Sprintf("/vrack/%s/cloudProject/%s", vrackId, params.Project)

	err := config.OVHClient.Delete(endpoint, &r)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Waiting for Attachment Deletion Task id %d: VRack %s ->  PublicCloud %s", r.Id, vrackId, params.Project)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"init", "todo", "doing"},
		Target:     []string{"completed"},
		Refresh:    VRackTaskRefreshFunc(config.OVHClient, vrackId, r.Id),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for vrack (%s) to attach to public cloud (%s): %s", vrackId, params.Project, err)
	}
	log.Printf("[DEBUG] Removed Attachement id %d: VRack %s ->  PublicCloud %s", r.Id, vrackId, params.Project)

	d.SetId("")
	return nil
}

func vrackPublicCloudAttachmentExists(vrackId, projectId string, c *ovh.Client) error {
	type attachResponse struct {
		VRack   string `json:"vrack"`
		Project string `json:"project"`
	}

	r := attachResponse{}

	endpoint := fmt.Sprintf("/vrack/%s/cloudProject/%s", vrackId, projectId)

	err := c.Get(endpoint, &r)
	if err != nil {
		return fmt.Errorf("Error while querying %s: %q\n", endpoint, err)
	}
	log.Printf("[DEBUG] Read Attachment %s -> VRack:%s, Cloud Project: %s", endpoint, r.VRack, r.Project)

	return nil
}

// AttachmentStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// an Attachment Task.
func VRackTaskRefreshFunc(c *ovh.Client, serviceName string, taskId int) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		r := attachTaskResponse{}
		endpoint := fmt.Sprintf("/vrack/%s/task/%d", serviceName, taskId)
		err := c.Get(endpoint, &r)
		if err != nil {
			if err.(*ovh.APIError).Code == 404 {
				log.Printf("[DEBUG] Task id %d on VRack %s completed", taskId, serviceName)
				return taskId, "completed", nil
			} else {
				return taskId, "", err
			}
		}

		log.Printf("[DEBUG] Pending Task id %d on VRack %s status: %s", r.Id, serviceName, r.Status)
		return taskId, r.Status, nil
	}
}
