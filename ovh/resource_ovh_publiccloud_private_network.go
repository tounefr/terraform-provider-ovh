package ovh

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ovh/go-ovh/ovh"
	"log"
	"os"
	"time"
)

func resourcePublicCloudPrivateNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourcePublicCloudPrivateNetworkCreate,
		Read:   resourcePublicCloudPrivateNetworkRead,
		Update: resourcePublicCloudPrivateNetworkUpdate,
		Delete: resourcePublicCloudPrivateNetworkDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("project_id", os.Getenv("OVH_PROJECT_ID"))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"project_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVH_PROJECT_ID", ""),
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"vlan_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  0,
			},
			"regions": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"regions_status": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"region": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},
			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// Params
type pcpnCreateParams struct {
	ProjectId string   `json:"serviceName"`
	VlanId    int      `json:"vlanId"`
	Name      string   `json:"name"`
	Regions   []string `json:"regions"`
}

func (p *pcpnCreateParams) String() string {
	return fmt.Sprintf("projectId: %s, vlanId:%d, name: %s, regions: %s", p.ProjectId, p.VlanId, p.Name, p.Regions)
}

// Params
type pcpnUpdateParams struct {
	Name string `json:"name"`
}

type pcpnRegion struct {
	Status string `json:"status"`
	Region string `json:"region"`
}

func (p *pcpnRegion) String() string {
	return fmt.Sprintf("Status:%s, Region: %s", p.Status, p.Region)
}

type pcpnResponse struct {
	Id      string        `json:"id"`
	Status  string        `json:"status"`
	Vlanid  int           `json:"vlanId"`
	Name    string        `json:"name"`
	Type    string        `json:"type"`
	Regions []*pcpnRegion `json:"regions"`
}

func (p *pcpnResponse) String() string {
	return fmt.Sprintf("Id: %s, Status: %s, Name: %s, Vlanid: %d, Type: %s, Regions: %s", p.Id, p.Status, p.Name, p.Vlanid, p.Type, p.Regions)
}

func regionsParamsFromSchema(d *schema.ResourceData) []string {
	var regions []string
	if v := d.Get("regions"); v != nil {
		rs := v.(*schema.Set).List()
		if len(rs) > 0 {
			for _, v := range v.(*schema.Set).List() {
				regions = append(regions, v.(string))
			}
		}
	}
	return regions
}

func resourcePublicCloudPrivateNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	projectId := d.Get("project_id").(string)
	params := &pcpnCreateParams{
		ProjectId: d.Get("project_id").(string),
		VlanId:    d.Get("vlan_id").(int),
		Name:      d.Get("name").(string),
		Regions:   regionsParamsFromSchema(d),
	}

	r := &pcpnResponse{}

	log.Printf("[DEBUG] Will create public cloud private network: %s", params)

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private", params.ProjectId)

	err := config.OVHClient.Post(endpoint, params, r)
	if err != nil {
		return fmt.Errorf("[ERROR] calling %s with params %s:\n\t %q", endpoint, params, err)
	}

	log.Printf("[DEBUG] Waiting for Private Network %s:", r)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"BUILDING"},
		Target:     []string{"ACTIVE"},
		Refresh:    pcpnRefreshFunc(config.OVHClient, projectId, r.Id),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("[ERROR] waiting for private network (%s): %s", params, err)
	}
	log.Printf("[DEBUG] Created Private Network %s", r)

	//set id
	d.SetId(r.Id)

	return nil
}

func resourcePublicCloudPrivateNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	projectId := d.Get("project_id").(string)
	params := &pcpnUpdateParams{
		Name: d.Get("name").(string),
	}

	log.Printf("[DEBUG] Will update public cloud private network: %s", params)

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s", projectId, d.Id())

	err := config.OVHClient.Put(endpoint, params, nil)
	if err != nil {
		return fmt.Errorf("[ERROR] calling %s with params %s:\n\t %q", endpoint, params, err)
	}

	log.Printf("[DEBUG] Updated Public cloud %s Private Network %s:", projectId, d.Id())

	return nil
}

func resourcePublicCloudPrivateNetworkRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	projectId := d.Get("project_id").(string)

	r := &pcpnResponse{}

	log.Printf("[DEBUG] Will read public cloud private network for project: %s, id: %s", projectId, d.Id())

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s", projectId, d.Id())

	err := config.OVHClient.Get(endpoint, r)
	if err != nil {
		return fmt.Errorf("[ERROR] calling %s:\n\t %q", endpoint, err)
	}

	readPcpn(d, r)

	log.Printf("[DEBUG] Read Public Cloud Private Network %s", r)
	return nil
}

func readPcpn(d *schema.ResourceData, r *pcpnResponse) {
	d.Set("name", r.Name)
	d.Set("status", r.Status)
	d.Set("type", r.Type)
	d.Set("vlan_id", r.Vlanid)

	regions_status := make([]map[string]interface{}, 0)
	regions := make([]string, 0)
	for i := range r.Regions {
		region := make(map[string]interface{})
		region["region"] = r.Regions[i].Region
		region["status"] = r.Regions[i].Status
		regions_status = append(regions_status, region)
		regions = append(regions, fmt.Sprintf(r.Regions[i].Region))
	}
	d.Set("regions_status", regions_status)
	d.Set("regions", regions)

	d.SetId(r.Id)
}

func resourcePublicCloudPrivateNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	projectId := d.Get("project_id").(string)
	id := d.Id()

	log.Printf("[DEBUG] Will delete public cloud private network for project: %s, id: %s", projectId, id)

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s", projectId, id)

	err := config.OVHClient.Delete(endpoint, nil)
	if err != nil {
		return fmt.Errorf("[ERROR] calling %s:\n\t %q", endpoint, err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"DELETING"},
		Target:     []string{"DELETED"},
		Refresh:    pcpnDelRefreshFunc(config.OVHClient, projectId, id),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("[ERROR] deleting for private network (%s): %s", id, err)
	}

	d.SetId("")

	log.Printf("[DEBUG] Deleted Public Cloud %s Private Network %s", projectId, id)
	return nil
}

func pcpnExists(projectId, id string, c *ovh.Client) error {
	r := &pcpnResponse{}

	log.Printf("[DEBUG] Will read public cloud private network for project: %s, id: %s", projectId, id)

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s", projectId, id)

	err := c.Get(endpoint, r)
	if err != nil {
		return fmt.Errorf("[ERROR] calling %s:\n\t %q", endpoint, err)
	}
	log.Printf("[DEBUG] Read public cloud private network: %s", r)

	return nil
}

// AttachmentStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// an Attachment Task.
func pcpnRefreshFunc(c *ovh.Client, projectId, pcpnId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		r := &pcpnResponse{}
		endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s", projectId, pcpnId)
		err := c.Get(endpoint, r)
		if err != nil {
			return r, "", err
		}

		log.Printf("[DEBUG] Pending Private Network: %s", r)
		return r, r.Status, nil
	}
}

// AttachmentStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// an Attachment Task.
func pcpnDelRefreshFunc(c *ovh.Client, projectId, pcpnId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		r := &pcpnResponse{}
		endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s", projectId, pcpnId)
		err := c.Get(endpoint, r)
		if err != nil {
			if err.(*ovh.APIError).Code == 404 {
				log.Printf("[DEBUG] private network id %s on project %s deleted", pcpnId, projectId)
				return r, "DELETED", nil
			} else {
				return r, "", err
			}
		}
		log.Printf("[DEBUG] Pending Private Network: %s", r)
		return r, r.Status, nil
	}
}
