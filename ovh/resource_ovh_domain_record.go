package ovh

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strconv"
)

func resourceDomainRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainRecordCreate,
		Read:   resourceDomainRecordRead,
		Update: resourceDomainRecordUpdate,
		Delete: resourceDomainRecordDelete,

		Schema: map[string]*schema.Schema{
			"domain": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"value": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"ttl": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVH_DOMAIN_DEFAULT_TTL", "3600"),
			},
		},
	}
}

type domainRecordCreateParams struct {
	FieldType string `json:"fieldType"`
	SubDomain string `json:"subDomain"`
	Target    string `json:"target"`
	TTL       string `json:"ttl"`
}

type domainRecordCreateResponse struct {
	Target    string `json:"target"`
	TTL       int    `json:"ttl"`
	Zone      string `json:"zone"`
	FieldType string `json:"fieldType"`
	Id        int    `json:"id"`
	SubDomain string `json:"subDomain"`
}

func resourceDomainRecordCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	// Delete existing entries to prevent multiple entries
	err, ids := domainRecordGetIds(d, meta)
	if err == nil {
		for id := range ids {
			d.SetId(strconv.Itoa(id))
			resourceDomainRecordDelete(d, meta)
		}
	}

	zoneName := d.Get("domain").(string)
	params := &domainRecordCreateParams{
		FieldType: d.Get("type").(string),
		SubDomain: d.Get("name").(string),
		Target:    d.Get("value").(string),
		TTL:       d.Get("ttl").(string),
	}

	log.Printf("[DEBUG] Will create domain record %s %s.%s to %s",
		params.FieldType, params.SubDomain, zoneName, params.Target)

	res := &domainRecordCreateResponse{}

	endpoint := fmt.Sprintf("/domain/zone/%s/record", zoneName)
	err = config.OVHClient.Post(endpoint, params, &res)
	if err != nil {
		return fmt.Errorf("[ERROR] calling Post %s with params %s:\n\t %q", endpoint, params, err)
	}
	d.SetId(strconv.Itoa(res.Id))
	log.Printf("[DEBUG] Domain record %s %s.%s to %s created",
		params.FieldType, params.SubDomain, zoneName, params.Target)

	domainRefresh(d, meta)
	return nil
}

type domainRecordGetParams struct {
	ZoneName  string `json:"zoneName"`
	SubDomain string `json:"subDomain"`
	FieldType string `json:"fieldType"`
}

func domainRefresh(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	endpoint := fmt.Sprintf("/domain/zone/%s/refresh")
	err := config.OVHClient.Post(endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("[ERROR] calling Post %s", endpoint)
	}
	return nil
}

func domainRecordGetIds(d *schema.ResourceData, meta interface{}) (error, []int) {
	config := meta.(*Config)

	params := domainRecordGetParams{
		ZoneName:  d.Get("domain").(string),
		SubDomain: d.Get("name").(string),
		FieldType: d.Get("type").(string),
	}
	log.Printf("[DEBUG] Will get domain record %s.%s", params.SubDomain, params.ZoneName)
	endpoint := fmt.Sprintf("/domain/zone/%s/record?subDomain=%s&fieldType=%s",
		params.ZoneName, params.SubDomain, params.FieldType)

	var ids []int
	err := config.OVHClient.Get(endpoint, &ids)
	if err != nil {
		return fmt.Errorf("[ERROR] calling Get %s with params %s:\n\t %q", endpoint, params, err), nil
	}
	if len(ids) == 0 {
		return fmt.Errorf("[ERROR] No domain record entries"), nil
	}
	log.Printf("[DEBUG] Domain record %s.%s fetched, ids: %v",
		params.SubDomain, params.ZoneName, ids)
	return nil, ids

}

type domainRecordReadResponse domainRecordCreateResponse

func resourceDomainRecordRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	err, ids := domainRecordGetIds(d, meta)
	if err != nil {
		return nil
	}
	id := strconv.Itoa(ids[0]) // In case of multiple entries for a record, i choose the first one

	zoneName := d.Get("name").(string)
	res := domainRecordReadResponse{}

	endpoint := fmt.Sprintf("/domain/zone/%s/record/%s", zoneName, id)
	err = config.OVHClient.Get(endpoint, &res)

	return nil
}

type domainRecordPutParams struct {
	Id        string `json:"id"`
	SubDomain string `json:"subDomain"`
	Target    string `json:"target"`
	TTL       string `json:"ttl"`
}

func resourceDomainRecordUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	zoneName := d.Get("domain").(string)
	params := domainRecordPutParams{
		Id:        d.Id(),
		SubDomain: d.Get("name").(string),
		Target:    d.Get("value").(string),
		TTL:       d.Get("ttl").(string),
	}

	endpoint := fmt.Sprintf("/domain/zone/%s/record/%s", zoneName, params.Id)
	err := config.OVHClient.Put(endpoint, params, nil)
	if err != nil {
		return fmt.Errorf("[ERROR] calling PUT %s:\n\t %q", endpoint, err)
	}

	domainRefresh(d, meta)
	return nil
}

type domainRecordDeleteParams struct {
	ZoneName  string `json:"zoneName"`
	Id        string `json:"id"`
	SubDomain string `json:"subDomain"`
	Target    string `json:"target"`
	TTL       string `json:"ttl"`
}

func resourceDomainRecordDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	params := domainRecordDeleteParams{
		ZoneName:  d.Get("domain").(string),
		Id:        d.Id(),
		SubDomain: d.Get("name").(string),
		Target:    d.Get("value").(string),
		TTL:       d.Get("ttl").(string),
	}

	log.Printf("[DEBUG] Will delete domain %s %s.%s to %s record",
		d.Get("type").(string), params.SubDomain, params.ZoneName, params.Target)

	endpoint := fmt.Sprintf("/domain/zone/%s/record/%s", params.ZoneName, params.Id)
	err := config.OVHClient.Delete(endpoint, params)
	if err != nil {
		return fmt.Errorf("[ERROR] calling DELETE %s:\n\t %q", endpoint, err)
	}

	fmt.Println("[DEBUG] Domain record deleted")
	d.SetId("")

	domainRefresh(d, meta)
	return nil
}
