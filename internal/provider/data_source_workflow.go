package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &WorkflowDataSource{}

func NewWorkflowDataSource() datasource.DataSource {
	return &WorkflowDataSource{}
}

type WorkflowDataSource struct {
	client *CircleCIClient
}

type WorkflowDataSourceModel struct {
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

// CircleCI API models for workflows
type WorkflowAPI struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PipelineID  string `json:"pipeline_id"`
	ProjectSlug string `json:"project_slug"`
	Status      string `json:"status"`
	StartedBy   string `json:"started_by"`
	CreatedAt   string `json:"created_at"`
	StoppedAt   string `json:"stopped_at"`
	Tag         string `json:"tag"`
}

func (d *WorkflowDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

func (d *WorkflowDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Workflow data source. This data source allows you to retrieve information about a specific workflow.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the workflow.",
				Required:            true,
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
				MarkdownDescription: "The current status of the workflow (success, running, not_run, failed, error, failing, on_hold, canceled, unauthorized).",
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
	}
}

func (d *WorkflowDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WorkflowDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WorkflowDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workflowID := data.ID.ValueString()
	if workflowID == "" {
		resp.Diagnostics.AddError("Missing Workflow ID", "The workflow ID is required")
		return
	}

	// Make API request to get workflow details using the client's MakeRequest method
	endpoint := fmt.Sprintf("/v2/workflow/%s", workflowID)
	httpResp, err := d.client.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to make request", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Failed to get workflow",
			fmt.Sprintf("CircleCI API returned status %d", httpResp.StatusCode),
		)
		return
	}

	var workflow WorkflowAPI
	if err := json.NewDecoder(httpResp.Body).Decode(&workflow); err != nil {
		resp.Diagnostics.AddError("Failed to decode response", err.Error())
		return
	}

	// Map API response to data source model
	data.ID = types.StringValue(workflow.ID)
	data.Name = types.StringValue(workflow.Name)
	data.PipelineID = types.StringValue(workflow.PipelineID)
	data.ProjectSlug = types.StringValue(workflow.ProjectSlug)
	data.Status = types.StringValue(workflow.Status)
	data.StartedBy = types.StringValue(workflow.StartedBy)
	data.CreatedAt = types.StringValue(workflow.CreatedAt)

	if workflow.StoppedAt != "" {
		data.StoppedAt = types.StringValue(workflow.StoppedAt)
	} else {
		data.StoppedAt = types.StringNull()
	}

	if workflow.Tag != "" {
		data.Tag = types.StringValue(workflow.Tag)
	} else {
		data.Tag = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
