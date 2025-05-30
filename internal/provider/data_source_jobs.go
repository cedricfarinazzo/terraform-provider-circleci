package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &JobsDataSource{}

func NewJobsDataSource() datasource.DataSource {
	return &JobsDataSource{}
}

// JobsDataSource defines the data source implementation.
type JobsDataSource struct {
	client *CircleCIClient
}

// JobsDataSourceModel describes the data source data model.
type JobsDataSourceModel struct {
	WorkflowID types.String         `tfsdk:"workflow_id"`
	Jobs       []JobDataSourceModel `tfsdk:"jobs"`
}

type JobDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ProjectSlug       types.String `tfsdk:"project_slug"`
	JobNumber         types.Int64  `tfsdk:"job_number"`
	Status            types.String `tfsdk:"status"`
	StartedAt         types.String `tfsdk:"started_at"`
	StoppedAt         types.String `tfsdk:"stopped_at"`
	ApprovalType      types.String `tfsdk:"approval_type"`
	ApprovalRequestID types.String `tfsdk:"approval_request_id"`
}

func (d *JobsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jobs"
}

func (d *JobsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about jobs within a specific CircleCI workflow.",

		Attributes: map[string]schema.Attribute{
			"workflow_id": schema.StringAttribute{
				MarkdownDescription: "The workflow ID for which to retrieve jobs.",
				Required:            true,
			},
			"jobs": schema.ListNestedAttribute{
				MarkdownDescription: "List of jobs in the workflow.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier of the job.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the job.",
							Computed:            true,
						},
						"project_slug": schema.StringAttribute{
							MarkdownDescription: "The project slug for the job.",
							Computed:            true,
						},
						"job_number": schema.Int64Attribute{
							MarkdownDescription: "The job number.",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The current status of the job.",
							Computed:            true,
						},
						"started_at": schema.StringAttribute{
							MarkdownDescription: "The date and time the job started.",
							Computed:            true,
						},
						"stopped_at": schema.StringAttribute{
							MarkdownDescription: "The date and time the job stopped.",
							Computed:            true,
						},
						"approval_type": schema.StringAttribute{
							MarkdownDescription: "The type of approval required (for approval jobs).",
							Computed:            true,
						},
						"approval_request_id": schema.StringAttribute{
							MarkdownDescription: "The approval request ID (for approval jobs).",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *JobsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (d *JobsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data JobsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Make API request to get jobs
	url := fmt.Sprintf("%s/workflow/%s/job", d.client.BaseURL, data.WorkflowID.ValueString())

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating request",
			fmt.Sprintf("Unable to create request for jobs: %v", err),
		)
		return
	}

	httpReq.Header.Set("Circle-Token", d.client.ApiToken)
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making request",
			fmt.Sprintf("Unable to get jobs: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Error from API",
			fmt.Sprintf("API returned status %d when getting jobs", httpResp.StatusCode),
		)
		return
	}

	var apiResponse struct {
		Items []struct {
			ID                string `json:"id"`
			Name              string `json:"name"`
			ProjectSlug       string `json:"project_slug"`
			JobNumber         int64  `json:"job_number"`
			Status            string `json:"status"`
			StartedAt         string `json:"started_at"`
			StoppedAt         string `json:"stopped_at"`
			ApprovalType      string `json:"approval_type"`
			ApprovalRequestID string `json:"approval_request_id"`
		} `json:"items"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error parsing response",
			fmt.Sprintf("Unable to parse jobs response: %v", err),
		)
		return
	}

	// Convert API response to model
	data.Jobs = make([]JobDataSourceModel, len(apiResponse.Items))
	for i, job := range apiResponse.Items {
		data.Jobs[i] = JobDataSourceModel{
			ID:                types.StringValue(job.ID),
			Name:              types.StringValue(job.Name),
			ProjectSlug:       types.StringValue(job.ProjectSlug),
			JobNumber:         types.Int64Value(job.JobNumber),
			Status:            types.StringValue(job.Status),
			StartedAt:         types.StringValue(job.StartedAt),
			StoppedAt:         types.StringValue(job.StoppedAt),
			ApprovalType:      types.StringValue(job.ApprovalType),
			ApprovalRequestID: types.StringValue(job.ApprovalRequestID),
		}
	}

	tflog.Trace(ctx, "read jobs data source", map[string]interface{}{
		"workflow_id": data.WorkflowID.ValueString(),
		"count":       len(data.Jobs),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
