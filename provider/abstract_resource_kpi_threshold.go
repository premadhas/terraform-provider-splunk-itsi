package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/tivo/terraform-provider-splunk-itsi/provider/util"
)

func getKpiThresholdPolicySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"policies": {
			Required:    true,
			Type:        schema.TypeSet,
			Description: "Map object of policies keyed by policy_name. ",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"policy_name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Internal key value for policy.",
					},
					"title": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The policy title, displayed to the user in the UI. Should be unique per policies object.",
					},
					"policy_type": {
						Type:     schema.TypeString,
						Required: true,
						Description: `The algorithm, specified for the current policy threshold level evaluation.
									   Supported values: static, stdev (standard deviation), quantile, range and percentage.`,
						ValidateFunc: validation.StringInSlice([]string{"static", "stdev", "quantile", "range", "percentage"}, false),
					},
					"time_blocks": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "Determines time periods which the policy is associated with.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"cron": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Corresponds to the cron expression in format: {minute} {hour} {\\*} {\\*} {day}",
								},
								"interval": {
									Type:        schema.TypeInt,
									Required:    true,
									Description: "The duration in minutes.",
								},
							},
						},
					},
					"aggregate_thresholds": {
						Required:    true,
						Type:        schema.TypeSet,
						Description: "User-defined thresholding levels for \"Aggregate\" threshold type. For more information, see KPI Threshold Setting.",
						Elem: &schema.Resource{
							Schema: getKpiThresholdSettingsSchema(),
						},
					},
					"entity_thresholds": {
						Required:    true,
						Type:        schema.TypeSet,
						Description: "User-defined thresholding levels for \"Per Entity\" threshold type. For more information, see KPI Threshold Setting.",
						Elem: &schema.Resource{
							Schema: getKpiThresholdSettingsSchema(),
						},
					},
				},
			},
		},
	}
}

func getKpiThresholdSettingsSchema() map[string]*schema.Schema {
	kpiThresholdLevel := map[string]*schema.Schema{
		"severity_label": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"info", "critical", "high", "medium", "low", "normal"}, false),
			Description:  "Severity label assigned for this threshold level like info, warning, critical, etc",
		},
		"threshold_value": {
			Type:     schema.TypeFloat,
			Required: true,
			Description: `Value for the threshold field stats identifying this threshold level. 
				This is the key value that defines the levels for values derived from the KPI search metrics.`,
		},
		"dynamic_param": {
			Type:        schema.TypeFloat,
			Computed:    true,
			Optional:    true,
			Description: "Value of the dynamic parameter for adaptive thresholds",
		},
	}

	return map[string]*schema.Schema{
		"base_severity_label": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "normal",
			ValidateFunc: validation.StringInSlice([]string{"info", "critical", "high", "medium", "low", "normal"}, false),
			Description:  "Base severity label assigned for the threshold (info, normal, low, medium, high, critical). ",
		},
		"gauge_max": {
			Type:        schema.TypeFloat,
			Optional:    true,
			Description: "Maximum value for the threshold gauge specified by user",
		},
		"gauge_min": {
			Type:        schema.TypeFloat,
			Optional:    true,
			Description: "Minimum value for the threshold gauge specified by user.",
		},
		"is_max_static": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "True when maximum threshold value is a static value, false otherwise. ",
		},
		"is_min_static": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "True when min threshold value is a static value, false otherwise.",
		},
		"metric_field": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Thresholding field from the search.",
		},
		"render_boundary_max": {
			Type:        schema.TypeFloat,
			Required:    true,
			Description: "Upper bound value to use to render the graph for the thresholds.",
		},
		"render_boundary_min": {
			Type:        schema.TypeFloat,
			Required:    true,
			Description: "Lower bound value to use to render the graph for the thresholds.",
		},
		"search": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Generated search used to compute the thresholds for this KPI.",
		},
		"threshold_levels": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: kpiThresholdLevel,
			},
		},
	}
}

func kpiThresholdSettingsToResourceData(sourceThresholdSetting map[string]interface{}, settingType string) (interface{}, error) {
	thresholdSetting := map[string]interface{}{}
	thresholdSetting["base_severity_label"] = sourceThresholdSetting["baseSeverityLabel"]
	thresholdSetting["gauge_max"] = sourceThresholdSetting["gaugeMax"]
	thresholdSetting["gauge_min"] = sourceThresholdSetting["gaugeMin"]
	thresholdSetting["is_max_static"] = sourceThresholdSetting["isMaxStatic"]
	thresholdSetting["is_min_static"] = sourceThresholdSetting["isMinStatic"]
	thresholdSetting["metric_field"] = sourceThresholdSetting["metricField"]
	thresholdSetting["render_boundary_max"] = sourceThresholdSetting["renderBoundaryMax"]
	thresholdSetting["render_boundary_min"] = sourceThresholdSetting["renderBoundaryMin"]
	thresholdSetting["search"] = sourceThresholdSetting["search"]
	thresholdLevels := []interface{}{}
	for _, tData_ := range sourceThresholdSetting["thresholdLevels"].([]interface{}) {
		tData := tData_.(map[string]interface{})
		thresholdLevel := map[string]interface{}{}
		switch tData["dynamicParam"] {
		case "":
			if settingType != "static" {
				return nil, fmt.Errorf("empty dynamic param for adaptive policy %s", settingType)
			}
			thresholdLevel["dynamic_param"] = 0
		default:
			thresholdLevel["dynamic_param"] = tData["dynamicParam"]
		}

		thresholdLevel["severity_label"] = tData["severityLabel"]
		thresholdLevel["threshold_value"] = tData["thresholdValue"]
		thresholdLevels = append(thresholdLevels, thresholdLevel)
	}
	thresholdSetting["threshold_levels"] = thresholdLevels
	return []interface{}{thresholdSetting}, nil
}

func kpiThresholdPolicyToResourceData(sourcePolicy map[string]interface{}, policyName string) (interface{}, error) {
	policy := map[string]interface{}{}
	policy["policy_name"] = policyName
	policy["title"] = sourcePolicy["title"]
	policy["policy_type"] = sourcePolicy["policy_type"]
	tfTimeBlocks := []interface{}{}
	for _, timeBlock := range sourcePolicy["time_blocks"].([]interface{}) {
		_timeBlock := timeBlock.([]interface{})
		tfTimeBlock := map[string]interface{}{
			"cron":     _timeBlock[0],
			"interval": _timeBlock[1],
		}
		tfTimeBlocks = append(tfTimeBlocks, tfTimeBlock)
	}
	policy["time_blocks"] = tfTimeBlocks

	var err error
	policy["aggregate_thresholds"], err =
		kpiThresholdSettingsToResourceData(sourcePolicy["aggregate_thresholds"].(map[string]interface{}), policy["policy_type"].(string))
	if err != nil {
		return nil, err
	}
	policy["entity_thresholds"], err =
		kpiThresholdSettingsToResourceData(sourcePolicy["entity_thresholds"].(map[string]interface{}), policy["policy_type"].(string))
	if err != nil {
		return nil, err
	}
	return policy, nil
}

func kpiThresholdPolicyToPayload(sourcePolicy map[string]interface{}) (interface{}, error) {
	policy := map[string]interface{}{}
	policy["title"] = sourcePolicy["title"].(string)
	policy["policy_type"] = sourcePolicy["policy_type"].(string)
	timeBlocks := [][]interface{}{}
	for _, b_ := range sourcePolicy["time_blocks"].(*schema.Set).List() {
		b := b_.(map[string]interface{})
		block := []interface{}{}
		block = append(block, b["cron"].(string))
		block = append(block, b["interval"].(int))

		timeBlocks = append(timeBlocks, block)
	}
	policy["time_blocks"] = timeBlocks
	for _, sourceAggregateThresholds := range sourcePolicy["aggregate_thresholds"].(*schema.Set).List() {
		aggregateThresholds, err := kpiThresholdThresholdSettingsToPayload(sourceAggregateThresholds.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		policy["aggregate_thresholds"] = aggregateThresholds
	}
	for _, sourceEntityThresholds := range sourcePolicy["entity_thresholds"].(*schema.Set).List() {
		entityThresholds, err := kpiThresholdThresholdSettingsToPayload(sourceEntityThresholds.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		policy["entity_thresholds"] = entityThresholds
	}
	return policy, nil
}

func kpiThresholdThresholdSettingsToPayload(source map[string]interface{}) (interface{}, error) {
	thresholdSetting := map[string]interface{}{}
	if severity, ok := util.SeverityMap[source["base_severity_label"].(string)]; ok {
		thresholdSetting["baseSeverityColor"] = severity.SeverityColor
		thresholdSetting["baseSeverityColorLight"] = severity.SeverityColorLight
		thresholdSetting["baseSeverityLabel"] = severity.SeverityLabel
		thresholdSetting["baseSeverityValue"] = severity.SeverityValue
	} else {
		return nil, fmt.Errorf("schema Validation broken. Unknown severity %s", source["base_severity_label"])
	}
	thresholdSetting["gaugeMax"] = source["gauge_max"].(float64)
	thresholdSetting["gaugeMin"] = source["gauge_min"].(float64)
	thresholdSetting["isMaxStatic"] = source["is_max_static"].(bool)
	thresholdSetting["isMinStatic"] = source["is_min_static"].(bool)
	thresholdSetting["metricField"] = source["metric_field"].(string)
	thresholdSetting["renderBoundaryMax"] = source["render_boundary_max"].(float64)
	thresholdSetting["renderBoundaryMin"] = source["render_boundary_min"].(float64)
	thresholdSetting["search"] = source["search"].(string)
	thresholdLevels := []interface{}{}
	for _, sourceThresholdLevel_ := range source["threshold_levels"].(*schema.Set).List() {
		sourceThresholdLevel := sourceThresholdLevel_.(map[string]interface{})
		thresholdLevel := map[string]interface{}{}
		thresholdLevel["dynamicParam"] = sourceThresholdLevel["dynamic_param"].(float64)
		if severity, ok := util.SeverityMap[sourceThresholdLevel["severity_label"].(string)]; ok {
			thresholdLevel["severityColor"] = severity.SeverityColor
			thresholdLevel["severityColorLight"] = severity.SeverityColorLight
			thresholdLevel["severityLabel"] = severity.SeverityLabel
			thresholdLevel["severityValue"] = severity.SeverityValue
		} else {
			return nil, fmt.Errorf("schema Validation broken. Unknown severity %s", sourceThresholdLevel["severity_label"])
		}
		thresholdLevel["thresholdValue"] = sourceThresholdLevel["threshold_value"].(float64)
		thresholdLevels = append(thresholdLevels, thresholdLevel)
	}
	thresholdSetting["thresholdLevels"] = thresholdLevels
	return thresholdSetting, nil
}
