package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &WorkflowsDataSource{}

func NewWorkflowsDataSource() datasource.DataSource {
	return &WorkflowsDataSource{}
}

type WorkflowsDataSource struct {
	client *CircleCIClient
}

type WorkflowsDataSourceModel struct {
	PipelineID types.String    `tfsdk:"pipeline_id"`
	Workflows  []WorkflowModel `tfsdk:"workflows"`
}

type WorkflowModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	PipelineID  types.String `tfsdk:"pipeline_id"`
	ProjectSlug types.String `tfsdk:"project_slug"`
	Status      types.String `tfsdk:"status"`
	StartedBy   types.String `tfsdk:"started_by"`
	CreatedAt   types.String `tfsdk:"created_at"`
	StoppedAt   types.String `tfsdk:"stopped_at"`
	Tag         types.String `tfsdk:"tag"`
}

// CircleCI API models for workflows list
type WorkflowsListAPI struct {
	Items         []WorkflowAPI `json:"items"`
	NextPageToken string        `json:"next_page_token"`
}

func (d *WorkflowsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflows"
}

func (d *WorkflowsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Workflows data source. This data source allows you to retrieve information about workflows in a specific pipeline.",

		Attributes: map[string]schema.Attribute{
			"pipeline_id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the pipeline to get workflows for.",
				Required:            true,
			},
			"workflows": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of workflows in the pipeline.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the workflow.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the workflow.",
						},
						"pipeline_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the pipeline that contains this workflow.",
						},
						"project_slug": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The project slug in the form 'vcs-slug/org-name/repo-name'.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The current status of the workflow.",
						},
						"started_by": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The user ID who started the workflow.",
						},
						"created_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The date and time the workflow was created.",
						},
						"stopped_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The date and time the workflow stopped running.",
						},
						"tag": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Tag used for the workflow.",
						},
					},
				},
			},
		},
	}
}

func (d *WorkflowsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WorkflowsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WorkflowsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pipelineID := data.PipelineID.ValueString()
	if pipelineID == "" {
		resp.Diagnostics.AddError("Missing Pipeline ID", "The pipeline ID is required")
		return
	}

	// Make API request to get workflows for the pipeline
	endpoint := fmt.Sprintf("/v2/pipeline/%s/workflow", pipelineID)
	httpResp, err := d.client.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to make request", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		resp.Diagnostics.AddError(
			"Failed to get workflows",
			fmt.Sprintf("CircleCI API returned status %d", httpResp.StatusCode),
		)
		return
	}

	var workflowsList WorkflowsListAPI
	if err := json.NewDecoder(httpResp.Body).Decode(&workflowsList); err != nil {
		resp.Diagnostics.AddError("Failed to decode response", err.Error())
		return
	}

	// Map API response to data source model
	data.PipelineID = types.StringValue(pipelineID)
	data.Workflows = make([]WorkflowModel, len(workflowsList.Items))

	for i, workflow := range workflowsList.Items {
		data.Workflows[i] = WorkflowModel{
			ID:          types.StringValue(workflow.ID),
			Name:        types.StringValue(workflow.Name),
			PipelineID:  types.StringValue(workflow.PipelineID),
			ProjectSlug: types.StringValue(workflow.ProjectSlug),
			Status:      types.StringValue(workflow.Status),
			StartedBy:   types.StringValue(workflow.StartedBy),
			CreatedAt:   types.StringValue(workflow.CreatedAt),
		}

		if workflow.StoppedAt != "" {
			data.Workflows[i].StoppedAt = types.StringValue(workflow.StoppedAt)
		} else {
			data.Workflows[i].StoppedAt = types.StringNull()
		}

		if workflow.Tag != "" {
			data.Workflows[i].Tag = types.StringValue(workflow.Tag)
		} else {
			data.Workflows[i].Tag = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
