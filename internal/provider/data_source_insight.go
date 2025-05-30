package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &InsightDataSource{}

func NewInsightDataSource() datasource.DataSource {
	return &InsightDataSource{}
}

type InsightDataSource struct {
	client *CircleCIClient
}

type InsightDataSourceModel struct {
	ProjectSlug types.String `tfsdk:"project_slug"`
	Branch      types.String `tfsdk:"branch"`
	Workflow    types.String `tfsdk:"workflow"`
	Metrics     types.Object `tfsdk:"metrics"`
}

type InsightMetrics struct {
	TotalRuns      types.Int64   `tfsdk:"total_runs"`
	SuccessfulRuns types.Int64   `tfsdk:"successful_runs"`
	MedianDuration types.Float64 `tfsdk:"median_duration"`
	P95Duration    types.Float64 `tfsdk:"p95_duration"`
	SuccessRate    types.Float64 `tfsdk:"success_rate"`
	Throughput     types.Float64 `tfsdk:"throughput"`
}

// CircleCI API models for insights
type InsightsResponse struct {
	TotalRuns      int     `json:"total_runs"`
	SuccessfulRuns int     `json:"successful_runs"`
	MedianDuration float64 `json:"median_duration_sec"`
	P95Duration    float64 `json:"p95_duration_sec"`
	SuccessRate    float64 `json:"success_rate"`
	Throughput     float64 `json:"throughput"`
}

func (d *InsightDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_insight"
}

func (d *InsightDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Insight data source. Use this data source to get insights and metrics about your workflows.",

		Attributes: map[string]schema.Attribute{
			"project_slug": schema.StringAttribute{
				MarkdownDescription: "The project slug in the form 'vcs-slug/org-name/repo-name'.",
				Required:            true,
			},
			"branch": schema.StringAttribute{
				MarkdownDescription: "The branch name to get insights for. Defaults to the default branch.",
				Optional:            true,
			},
			"workflow": schema.StringAttribute{
				MarkdownDescription: "The workflow name to get insights for. If not specified, gets insights for all workflows.",
				Optional:            true,
			},
			"metrics": schema.SingleNestedAttribute{
				MarkdownDescription: "The workflow metrics and insights.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"total_runs": schema.Int64Attribute{
						MarkdownDescription: "The total number of workflow runs.",
						Computed:            true,
					},
					"successful_runs": schema.Int64Attribute{
						MarkdownDescription: "The number of successful workflow runs.",
						Computed:            true,
					},
					"median_duration": schema.Float64Attribute{
						MarkdownDescription: "The median duration of workflow runs in seconds.",
						Computed:            true,
					},
					"p95_duration": schema.Float64Attribute{
						MarkdownDescription: "The 95th percentile duration of workflow runs in seconds.",
						Computed:            true,
					},
					"success_rate": schema.Float64Attribute{
						MarkdownDescription: "The success rate of workflow runs (0.0 to 1.0).",
						Computed:            true,
					},
					"throughput": schema.Float64Attribute{
						MarkdownDescription: "The average number of workflow runs per day.",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (d *InsightDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*CircleCIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *CircleCIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *InsightDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InsightDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := EscapeProjectSlug(data.ProjectSlug.ValueString())

	// Build the endpoint with optional parameters
	endpoint := fmt.Sprintf("/insights/pages/%s/summary", slug)
	params := make(map[string]string)

	if !data.Branch.IsNull() && !data.Branch.IsUnknown() {
		params["branch"] = data.Branch.ValueString()
	}

	if !data.Workflow.IsNull() && !data.Workflow.IsUnknown() {
		params["workflow_name"] = data.Workflow.ValueString()
	}

	url := BuildURL(endpoint, params)

	var insights InsightsResponse
	if err := d.client.Get(ctx, url, &insights); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read insights, got error: %s", err))
		return
	}

	// Convert to metrics object
	metricsObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"total_runs":      types.Int64Type,
		"successful_runs": types.Int64Type,
		"median_duration": types.Float64Type,
		"p95_duration":    types.Float64Type,
		"success_rate":    types.Float64Type,
		"throughput":      types.Float64Type,
	}, InsightMetrics{
		TotalRuns:      types.Int64Value(int64(insights.TotalRuns)),
		SuccessfulRuns: types.Int64Value(int64(insights.SuccessfulRuns)),
		MedianDuration: types.Float64Value(insights.MedianDuration),
		P95Duration:    types.Float64Value(insights.P95Duration),
		SuccessRate:    types.Float64Value(insights.SuccessRate),
		Throughput:     types.Float64Value(insights.Throughput),
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Metrics = metricsObj

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
