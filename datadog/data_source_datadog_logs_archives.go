package datadog

import (
	"context"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogLogsArchives() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to list several existing logs archives for use in other resources.",
		ReadContext: dataSourceDatadogLogsArchivesRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				// Computed values
				"logs_archives": {
					Description: "List of logs archives",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Description: "The ID of the archive.",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"type": {
								Description: "The type of the archive.",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"attributes": {
								Description: "The attributes of the archive.",
								Type:        schema.TypeMap,
								Computed:    true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
			}
		},
	}
}

func dataSourceDatadogLogsArchivesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	logsArchives, httpresp, err := apiInstances.GetLogsArchivesApiV2().ListLogsArchives(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying log indexes")
	}

	diags := diag.Diagnostics{}
	tfLogsArchives := make([]map[string]interface{}, len(logsArchives.GetData()))
	for i, l := range logsArchives.GetData() {
		if err := utils.CheckForUnparsed(l); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("skipping logs archive with name: %s", l.GetAttributes().Name),
				Detail:   fmt.Sprintf("logs archive contains unparsed object: %v", err),
			})
			continue
		}

		attr := l.GetAttributes()
		tfLogsArchives[i] = map[string]interface{}{
			"id":   l.GetId(),
			"type": l.GetType(),
			"attributes": map[string]interface{}{
				"name":  attr.GetName(),
				"query": attr.GetQuery(),
			},
		}
	}
	if err := d.Set("logs_archives", tfLogsArchives); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("log-archives")

	return diags
}
