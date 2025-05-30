package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &JobResource{}
var _ resource.ResourceWithImportState = &JobResource{}

func NewJobResource() resource.Resource {
	return &JobResource{}
}

type JobResource struct {
	client *CircleCIClient
}

type JobResourceModel struct {
	ID           types.String `tfsdk:"id"`
	JobNumber    types.Int64  `tfsdk:"job_number"`
	Name         types.String `tfsdk:"name"`
	ProjectSlug  types.String `tfsdk:"project_slug"`
	Status       types.String `tfsdk:"status"`
	StartedAt    types.String `tfsdk:"started_at"`
	StoppedAt    types.String `tfsdk:"stopped_at"`
	ApprovalID   types.String `tfsdk:"approval_id"`
	ApprovalType types.String `tfsdk:"approval_type"`
	Action       types.String `tfsdk:"action"` // "cancel", "approve", "rerun"
}

// CircleCI API models for jobs
type JobAPI struct {
	ID           string `json:"id"`
	JobNumber    int64  `json:"job_number"`
	Name         string `json:"name"`
	ProjectSlug  string `json:"project_slug"`
	Status       string `json:"status"`
	StartedAt    string `json:"started_at"`
	StoppedAt    string `json:"stopped_at"`
	ApprovalID   string `json:"approval_request_id,omitempty"`
	ApprovalType string `json:"type,omitempty"`
}

func (r *JobResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job"
}

func (r *JobResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Job resource. This resource allows you to manage CircleCI jobs, including canceling, approving, or rerunning them.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the job.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"job_number": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The job number.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the job.",
			},
			"project_slug": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The project slug in the form 'vcs-slug/org-name/repo-name'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current status of the job.",
			},
			"started_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time the job started.",
			},
			"stopped_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time the job stopped.",
			},
			"approval_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The approval request ID (for approval jobs).",
			},
			"approval_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The type of approval required.",
			},
			"action": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Action to perform on the job: 'cancel', 'approve', or 'rerun'.",
			},
		},
	}
}

func (r *JobResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*CircleCIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *CircleCIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *JobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data JobResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// For job resource, "create" means getting the job details and optionally performing an action
	projectSlug := data.ProjectSlug.ValueString()
	jobNumber := data.JobNumber.ValueInt64()

	// Get job details
	endpoint := fmt.Sprintf("/v2/project/%s/job/%d", projectSlug, jobNumber)
	httpResp, err := r.client.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get job", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Failed to get job",
			fmt.Sprintf("CircleCI API returned status %d", httpResp.StatusCode),
		)
		return
	}

	var job JobAPI
	if err := json.NewDecoder(httpResp.Body).Decode(&job); err != nil {
		resp.Diagnostics.AddError("Failed to decode response", err.Error())
		return
	}

	// Perform action if specified
	action := data.Action.ValueString()
	if action != "" {
		if err := r.performJobAction(ctx, projectSlug, jobNumber, action, data.ApprovalID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to perform job action", err.Error())
			return
		}
	}

	// Map job data to model
	r.mapJobToModel(&job, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data JobResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get job details
	projectSlug := data.ProjectSlug.ValueString()
	jobNumber := data.JobNumber.ValueInt64()

	endpoint := fmt.Sprintf("/v2/project/%s/job/%d", projectSlug, jobNumber)
	httpResp, err := r.client.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get job", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Failed to get job",
			fmt.Sprintf("CircleCI API returned status %d", httpResp.StatusCode),
		)
		return
	}

	var job JobAPI
	if err := json.NewDecoder(httpResp.Body).Decode(&job); err != nil {
		resp.Diagnostics.AddError("Failed to decode response", err.Error())
		return
	}

	r.mapJobToModel(&job, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data JobResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if action has changed
	action := data.Action.ValueString()
	if action != "" {
		projectSlug := data.ProjectSlug.ValueString()
		jobNumber := data.JobNumber.ValueInt64()

		if err := r.performJobAction(ctx, projectSlug, jobNumber, action, data.ApprovalID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to perform job action", err.Error())
			return
		}
	}

	// Read updated state
	r.Read(ctx, resource.ReadRequest{State: req.State}, &resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics})
}

func (r *JobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Job resources are read-only in terms of deletion
	// The job will continue to exist in CircleCI regardless of Terraform state
}

func (r *JobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "project_slug:job_number"
	// Example: "gh/myorg/myrepo:123"
	idParts := strings.SplitN(req.ID, ":", 2)
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Import ID must be in format 'project_slug:job_number' (e.g., 'gh/myorg/myrepo:123'), got: %q", req.ID),
		)
		return
	}

	projectSlug := idParts[0]
	jobNumberStr := idParts[1]

	// Convert job number string to int64
	jobNumber, err := strconv.ParseInt(jobNumberStr, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid job number",
			fmt.Sprintf("Job number must be a valid integer, got: %q", jobNumberStr),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_slug"), projectSlug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_number"), jobNumber)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *JobResource) performJobAction(ctx context.Context, projectSlug string, jobNumber int64, action, approvalID string) error {
	switch action {
	case "cancel":
		endpoint := fmt.Sprintf("/v2/project/%s/job/%d/cancel", projectSlug, jobNumber)
		httpResp, err := r.client.MakeRequest(ctx, "POST", endpoint, nil)
		if err != nil {
			return err
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to cancel job, status: %d", httpResp.StatusCode)
		}

	case "approve":
		if approvalID == "" {
			return fmt.Errorf("approval_id is required for approve action")
		}
		endpoint := fmt.Sprintf("/v2/workflow/%s/approve/%s", projectSlug, approvalID)
		httpResp, err := r.client.MakeRequest(ctx, "POST", endpoint, nil)
		if err != nil {
			return err
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to approve job, status: %d", httpResp.StatusCode)
		}

	case "rerun":
		endpoint := fmt.Sprintf("/v2/project/%s/job/%d/rerun", projectSlug, jobNumber)
		httpResp, err := r.client.MakeRequest(ctx, "POST", endpoint, nil)
		if err != nil {
			return err
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to rerun job, status: %d", httpResp.StatusCode)
		}

	default:
		return fmt.Errorf("unsupported action: %s. Supported actions: cancel, approve, rerun", action)
	}

	return nil
}

func (r *JobResource) mapJobToModel(job *JobAPI, data *JobResourceModel) {
	data.ID = types.StringValue(job.ID)
	data.JobNumber = types.Int64Value(job.JobNumber)
	data.Name = types.StringValue(job.Name)
	data.ProjectSlug = types.StringValue(job.ProjectSlug)
	data.Status = types.StringValue(job.Status)
	data.StartedAt = types.StringValue(job.StartedAt)

	if job.StoppedAt != "" {
		data.StoppedAt = types.StringValue(job.StoppedAt)
	} else {
		data.StoppedAt = types.StringNull()
	}

	if job.ApprovalID != "" {
		data.ApprovalID = types.StringValue(job.ApprovalID)
	} else {
		data.ApprovalID = types.StringNull()
	}

	if job.ApprovalType != "" {
		data.ApprovalType = types.StringValue(job.ApprovalType)
	} else {
		data.ApprovalType = types.StringNull()
	}
}
