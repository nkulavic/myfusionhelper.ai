package templates

import apitypes "github.com/myfusionhelper/api/internal/types"

// allTemplates defines all available dashboard templates.
var allTemplates = []apitypes.DashboardTemplate{
	{
		ID:          "tpl-generic-crm-overview",
		Name:        "CRM Overview",
		Description: "General dashboard for any connected CRM platform",
		Platform:    "generic",
		Widgets: []apitypes.DashboardWidget{
			{WidgetID: "w1", Type: "scorecard", Title: "Total Contacts", DataSource: "contacts", Metric: "count", Dimension: "_total", Size: "sm", Order: 0},
			{WidgetID: "w2", Type: "scorecard", Title: "Total Tags", DataSource: "tags", Metric: "count", Dimension: "_total", Size: "sm", Order: 1},
			{WidgetID: "w3", Type: "bar", Title: "Contacts by Status", DataSource: "contacts", Metric: "count", Dimension: "status", Size: "md", Order: 2},
			{WidgetID: "w4", Type: "pie", Title: "Tags by Category", DataSource: "tags", Metric: "count", Dimension: "category", Size: "md", Order: 3},
			{WidgetID: "w5", Type: "line", Title: "New Contacts Over Time", DataSource: "contacts", Metric: "count", Dimension: "_timeseries", Size: "lg", Order: 4},
			{WidgetID: "w6", Type: "table", Title: "Top Tags by Contact Count", DataSource: "tags", Metric: "count", Dimension: "name", Size: "full", Order: 5},
		},
	},
	{
		ID:          "tpl-keap-dashboard",
		Name:        "Keap Dashboard",
		Description: "Dashboard optimized for Keap (Infusionsoft) data",
		Platform:    "keap",
		Widgets: []apitypes.DashboardWidget{
			{WidgetID: "w1", Type: "scorecard", Title: "Total Contacts", DataSource: "contacts", Metric: "count", Dimension: "_total", Size: "sm", Order: 0},
			{WidgetID: "w2", Type: "scorecard", Title: "Active Contacts", DataSource: "contacts", Metric: "count", Dimension: "status", MetricField: "active", Size: "sm", Order: 1},
			{WidgetID: "w3", Type: "bar", Title: "Contacts by Lead Source", DataSource: "contacts", Metric: "count", Dimension: "lead_source", Size: "md", Order: 2},
			{WidgetID: "w4", Type: "pie", Title: "Contacts by Status", DataSource: "contacts", Metric: "count", Dimension: "status", Size: "md", Order: 3},
			{WidgetID: "w5", Type: "bar", Title: "Contacts by Source", DataSource: "contacts", Metric: "count", Dimension: "source", Size: "md", Order: 4},
			{WidgetID: "w6", Type: "line", Title: "New Contacts Over Time", DataSource: "contacts", Metric: "count", Dimension: "_timeseries", Size: "lg", Order: 5},
			{WidgetID: "w7", Type: "table", Title: "Top Companies by Contact Count", DataSource: "contacts", Metric: "count", Dimension: "company", Size: "full", Order: 6},
		},
	},
	{
		ID:          "tpl-gohighlevel-dashboard",
		Name:        "GoHighLevel Dashboard",
		Description: "Dashboard optimized for GoHighLevel data",
		Platform:    "gohighlevel",
		Widgets: []apitypes.DashboardWidget{
			{WidgetID: "w1", Type: "scorecard", Title: "Total Contacts", DataSource: "contacts", Metric: "count", Dimension: "_total", Size: "sm", Order: 0},
			{WidgetID: "w2", Type: "bar", Title: "Contacts by Source", DataSource: "contacts", Metric: "count", Dimension: "source", Size: "md", Order: 1},
			{WidgetID: "w3", Type: "pie", Title: "Contacts by Status", DataSource: "contacts", Metric: "count", Dimension: "status", Size: "md", Order: 2},
			{WidgetID: "w4", Type: "line", Title: "New Contacts Over Time", DataSource: "contacts", Metric: "count", Dimension: "_timeseries", Size: "lg", Order: 3},
			{WidgetID: "w5", Type: "table", Title: "Top Tags Applied", DataSource: "tags", Metric: "count", Dimension: "name", Size: "full", Order: 4},
		},
	},
	{
		ID:          "tpl-activecampaign-dashboard",
		Name:        "ActiveCampaign Dashboard",
		Description: "Dashboard optimized for ActiveCampaign data",
		Platform:    "activecampaign",
		Widgets: []apitypes.DashboardWidget{
			{WidgetID: "w1", Type: "scorecard", Title: "Total Contacts", DataSource: "contacts", Metric: "count", Dimension: "_total", Size: "sm", Order: 0},
			{WidgetID: "w2", Type: "scorecard", Title: "Total Tags", DataSource: "tags", Metric: "count", Dimension: "_total", Size: "sm", Order: 1},
			{WidgetID: "w3", Type: "bar", Title: "Contacts by Status", DataSource: "contacts", Metric: "count", Dimension: "status", Size: "md", Order: 2},
			{WidgetID: "w4", Type: "pie", Title: "Tags by Category", DataSource: "tags", Metric: "count", Dimension: "category", Size: "md", Order: 3},
			{WidgetID: "w5", Type: "line", Title: "New Contacts Over Time", DataSource: "contacts", Metric: "count", Dimension: "_timeseries", Size: "lg", Order: 4},
			{WidgetID: "w6", Type: "table", Title: "Top Tags by Contact Count", DataSource: "tags", Metric: "count", Dimension: "name", Size: "full", Order: 5},
		},
	},
	{
		ID:          "tpl-ontraport-dashboard",
		Name:        "Ontraport Dashboard",
		Description: "Dashboard optimized for Ontraport data",
		Platform:    "ontraport",
		Widgets: []apitypes.DashboardWidget{
			{WidgetID: "w1", Type: "scorecard", Title: "Total Contacts", DataSource: "contacts", Metric: "count", Dimension: "_total", Size: "sm", Order: 0},
			{WidgetID: "w2", Type: "bar", Title: "Contacts by Status", DataSource: "contacts", Metric: "count", Dimension: "status", Size: "md", Order: 1},
			{WidgetID: "w3", Type: "line", Title: "New Contacts Over Time", DataSource: "contacts", Metric: "count", Dimension: "_timeseries", Size: "lg", Order: 2},
			{WidgetID: "w4", Type: "table", Title: "Top Tags Applied", DataSource: "tags", Metric: "count", Dimension: "name", Size: "full", Order: 3},
		},
	},
	{
		ID:          "tpl-hubspot-dashboard",
		Name:        "HubSpot Dashboard",
		Description: "Dashboard optimized for HubSpot data",
		Platform:    "hubspot",
		Widgets: []apitypes.DashboardWidget{
			{WidgetID: "w1", Type: "scorecard", Title: "Total Contacts", DataSource: "contacts", Metric: "count", Dimension: "_total", Size: "sm", Order: 0},
			{WidgetID: "w2", Type: "bar", Title: "Contacts by Source", DataSource: "contacts", Metric: "count", Dimension: "source", Size: "md", Order: 1},
			{WidgetID: "w3", Type: "pie", Title: "Contacts by Status", DataSource: "contacts", Metric: "count", Dimension: "status", Size: "md", Order: 2},
			{WidgetID: "w4", Type: "line", Title: "New Contacts Over Time", DataSource: "contacts", Metric: "count", Dimension: "_timeseries", Size: "lg", Order: 3},
			{WidgetID: "w5", Type: "table", Title: "Top Companies by Contact Count", DataSource: "contacts", Metric: "count", Dimension: "company", Size: "full", Order: 4},
		},
	},
}

// GetAllTemplates returns all available templates.
func GetAllTemplates() []apitypes.DashboardTemplate {
	return allTemplates
}

// GetTemplateByID returns a template by ID, or nil if not found.
func GetTemplateByID(id string) *apitypes.DashboardTemplate {
	for i := range allTemplates {
		if allTemplates[i].ID == id {
			return &allTemplates[i]
		}
	}
	return nil
}

// GetTemplatesForPlatforms returns templates matching the given platform slugs plus generic templates.
func GetTemplatesForPlatforms(platformSlugs []string) []apitypes.DashboardTemplate {
	slugSet := make(map[string]bool, len(platformSlugs))
	for _, s := range platformSlugs {
		slugSet[s] = true
	}

	var result []apitypes.DashboardTemplate
	for _, t := range allTemplates {
		if t.Platform == "generic" || slugSet[t.Platform] {
			result = append(result, t)
		}
	}
	return result
}
