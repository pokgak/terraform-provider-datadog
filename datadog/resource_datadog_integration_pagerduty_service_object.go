package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const maskedSecret = "*****"

func resourceDatadogIntegrationPagerdutySO() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides access to individual Service Objects of Datadog - PagerDuty integrations. Note that the Datadog - PagerDuty integration must be activated in the Datadog UI in order for this resource to be usable.",
		CreateContext: resourceDatadogIntegrationPagerdutySOCreate,
		ReadContext:   resourceDatadogIntegrationPagerdutySORead,
		UpdateContext: resourceDatadogIntegrationPagerdutySOUpdate,
		DeleteContext: resourceDatadogIntegrationPagerdutySODelete,
		// since the API never returns service_key, it's impossible to meaningfully import resources
		Importer: nil,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"service_name": {
					Description: "Your Service name in PagerDuty.",
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
				},
				"service_key": {
					Description: "Your Service name associated service key in PagerDuty. This key may also be referred to as an Integration Key or Routing Key in the Pagerduty Integration [documentation](https://www.pagerduty.com/docs/guides/datadog-integration-guide/), UI, and within the [Pagerduty Provider for Terraform](https://registry.terraform.io/providers/PagerDuty/pagerduty/latest/docs/resources/service_integration#integration_key) Note: Since the Datadog API never returns service keys, it is impossible to detect [drifts](https://www.hashicorp.com/blog/detecting-and-managing-drift-with-terraform). The best way to solve a drift is to manually mark the Service Object resource with [terraform taint](https://www.terraform.io/docs/commands/taint.html) to have it destroyed and recreated.",
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
				},
			}
		},
	}
}

func buildIntegrationPagerdutySO(d *schema.ResourceData) *datadogV1.PagerDutyService {
	so := &datadogV1.PagerDutyService{}
	if v, ok := d.GetOk("service_name"); ok {
		so.SetServiceName(v.(string))
	}
	if v, ok := d.GetOk("service_key"); ok {
		so.SetServiceKey(v.(string))
	}

	return so
}

func buildIntegrationPagerdutyServiceKey(d *schema.ResourceData) *datadogV1.PagerDutyServiceKey {
	serviceKey := &datadogV1.PagerDutyServiceKey{}
	if v, ok := d.GetOk("service_key"); ok {
		serviceKey.SetServiceKey(v.(string))
	}

	return serviceKey
}

func resourceDatadogIntegrationPagerdutySOCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	so := buildIntegrationPagerdutySO(d)
	if _, httpresp, err := apiInstances.GetPagerDutyIntegrationApiV1().CreatePagerDutyIntegrationService(auth, *so); err != nil {
		// TODO: warn user that PD integration must be enabled to be able to create service objects
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating PagerDuty integration service")
	}
	d.SetId(so.GetServiceName())

	return resourceDatadogIntegrationPagerdutySORead(ctx, d, meta)
}

func resourceDatadogIntegrationPagerdutySORead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	so, httpresp, err := apiInstances.GetPagerDutyIntegrationApiV1().GetPagerDutyIntegrationService(auth, d.Id())
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting PagerDuty integration service")
	}
	if err := utils.CheckForUnparsed(so); err != nil {
		return diag.FromErr(err)
	}

	d.Set("service_name", so.GetServiceName())
	// Only update service_key if not set on d - the API endpoints never return
	// the keys, so this is how we recognize new values.
	if _, ok := d.GetOk("service_key"); !ok {
		d.Set("service_key", maskedSecret)
	}

	return nil
}

func resourceDatadogIntegrationPagerdutySOUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	serviceKey := buildIntegrationPagerdutyServiceKey(d)
	if httpresp, err := apiInstances.GetPagerDutyIntegrationApiV1().UpdatePagerDutyIntegrationService(auth, d.Id(), *serviceKey); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating PagerDuty integration service")
	}

	return resourceDatadogIntegrationPagerdutySORead(ctx, d, meta)
}

func resourceDatadogIntegrationPagerdutySODelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	if httpresp, err := apiInstances.GetPagerDutyIntegrationApiV1().DeletePagerDutyIntegrationService(auth, d.Id()); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting PagerDuty integration service")
	}

	return nil
}
