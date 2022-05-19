package resources

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tivo/terraform-provider-splunk-itsi/models"
)

func DatasourceKPIThresholdTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: DatasourceKPIThresholdTemplateRead,
		Schema: map[string]*schema.Schema{
			"title": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the KPI Threshold template",
			},
		},
	}
}

func DatasourceKPIThresholdTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	base := kpiThresholdTemplateBase(m.(models.ClientConfig), d.Id(), d.Get("title").(string))
	b, err := base.Find(ctx)
	if err != nil {
		return append(diags, diag.Errorf("%s", err)...)
	}
	if b == nil {
		d.SetId("")
		return nil
	}

	err = populateKPIThresholdTemplateDatasourceData(b, d)
	if err != nil {
		return append(diags, diag.Errorf("%s", err)...)
	}
	return
}

func populateKPIThresholdTemplateDatasourceData(b *models.Base, d *schema.ResourceData) error {
	by, err := b.RawJson.MarshalJSON()
	if err != nil {
		return err
	}
	var interfaceMap map[string]interface{}
	err = json.Unmarshal(by, &interfaceMap)
	if err != nil {
		return err
	}

	err = d.Set("title", interfaceMap["title"])
	if err != nil {
		return err
	}

	d.SetId(b.RESTKey)
	return nil
}
