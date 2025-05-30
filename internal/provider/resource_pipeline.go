package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &PipelineResource{}
var _ resource.ResourceWithImportState = &PipelineResource{}

func NewPipelineResource() resource.Resource {
	return &PipelineResource{}
}

type PipelineResource struct {
	client *CircleCIClient
}

type PipelineResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectSlug types.String `tfsdk:"project_slug"`
	Branch      types.String `tfsdk:"branch"`
	Tag         types.String `tfsdk:"tag"`
	Parameters  types.Map    `tfsdk:"parameters"`
	Number      types.Int64  `tfsdk:"number"`
	State       types.String `tfsdk:"state"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	VCS         types.Object `tfsdk:"vcs"`
}

type PipelineVCS struct {
	ProviderName        types.String `tfsdk:"provider_name"`
	TargetRepositoryURL types.String `tfsdk:"target_repository_url"`
	Branch              types.String `tfsdk:"branch"`
	ReviewID            types.String `tfsdk:"review_id"`
	ReviewURL           types.String `tfsdk:"review_url"`
	Revision            types.String `tfsdk:"revision"`
	Tag                 types.String `tfsdk:"tag"`
	Commit              types.Object `tfsdk:"commit"`
}

type PipelineCommit struct {
	Subject types.String `tfsdk:"subject"`
	Body    types.String `tfsdk:"body"`
}

// CircleCI API models for pipelines
type Pipeline struct {
	ID          string                 `json:"id"`
	Number      int64                  `json:"number"`
	ProjectSlug string                 `json:"project_slug"`
	State       string                 `json:"state"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
	VCS         VCSInfo                `json:"vcs"`
	Parameters  map[string]interface{} `json:"pipeline_parameters,omitempty"`
}

type VCSInfo struct {
	ProviderName        string     `json:"provider_name"`
	TargetRepositoryURL string     `json:"target_repository_url"`
	Branch              string     `json:"branch"`
	ReviewID            string     `json:"review_id"`
	ReviewURL           string     `json:"review_url"`
	Revision            string     `json:"revision"`
	Tag                 string     `json:"tag"`
	Commit              CommitInfo `json:"commit"`
}

type CommitInfo struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type TriggerPipelineRequest struct {
	Branch     string                 `json:"branch,omitempty"`
	Tag        string                 `json:"tag,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

func (r *PipelineResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline"
}

func (r *PipelineResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Pipeline resource. This resource allows you to trigger pipeline runs programmatically.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the pipeline.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_slug": schema.StringAttribute{
				MarkdownDescription: "The project slug in the form 'vcs-slug/org-name/repo-name'.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"branch": schema.StringAttribute{
				MarkdownDescription: "The branch to run the pipeline on. Cannot be used with tag.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tag": schema.StringAttribute{
				MarkdownDescription: "The tag to run the pipeline on. Cannot be used with branch.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"parameters": schema.MapAttribute{
				MarkdownDescription: "Pipeline parameters to pass to the pipeline.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers:       []planmodifier.Map{
					// Require replacement since pipelines are immutable once triggered
				},
			},
			"number": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The pipeline number.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current state of the pipeline.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time when the pipeline was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time when the pipeline was last updated.",
			},
			"vcs": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "VCS information for the pipeline.",
				Attributes: map[string]schema.Attribute{
					"provider_name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The VCS provider name.",
					},
					"target_repository_url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The target repository URL.",
					},
					"branch": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The branch name.",
					},
					"review_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The review ID.",
					},
					"review_url": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The review URL.",
					},
					"revision": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The git revision.",
					},
					"tag": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The tag name.",
					},
					"commit": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Commit information.",
						Attributes: map[string]schema.Attribute{
							"subject": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "The commit subject.",
							},
							"body": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "The commit body.",
							},
						},
					},
				},
			},
		},
	}
}

func (r *PipelineResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PipelineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PipelineResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either branch or tag is specified, but not both
	if data.Branch.IsNull() && data.Tag.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("branch"),
			"Missing Required Attribute",
			"Either 'branch' or 'tag' must be specified.",
		)
		return
	}

	if !data.Branch.IsNull() && !data.Tag.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("branch"),
			"Conflicting Attributes",
			"Cannot specify both 'branch' and 'tag'.",
		)
		return
	}

	// Extract parameters
	var parameters map[string]string
	if !data.Parameters.IsNull() {
		resp.Diagnostics.Append(data.Parameters.ElementsAs(ctx, &parameters, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	triggerReq := TriggerPipelineRequest{}
	if !data.Branch.IsNull() {
		triggerReq.Branch = data.Branch.ValueString()
	}
	if !data.Tag.IsNull() {
		triggerReq.Tag = data.Tag.ValueString()
	}

	if parameters != nil {
		triggerReq.Parameters = make(map[string]interface{})
		for k, v := range parameters {
			triggerReq.Parameters[k] = v
		}
	}

	slug := EscapeProjectSlug(data.ProjectSlug.ValueString())
	endpoint := fmt.Sprintf("/project/%s/pipeline", slug)

	var pipeline Pipeline
	if err := r.client.Post(ctx, endpoint, triggerReq, &pipeline); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to trigger pipeline, got error: %s", err))
		return
	}

	// Update the model with the response
	data.ID = types.StringValue(pipeline.ID)
	data.Number = types.Int64Value(pipeline.Number)
	data.State = types.StringValue(pipeline.State)
	data.CreatedAt = types.StringValue(pipeline.CreatedAt)
	data.UpdatedAt = types.StringValue(pipeline.UpdatedAt)

	// Convert VCS info to types.Object
	commitObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"subject": types.StringType,
		"body":    types.StringType,
	}, PipelineCommit{
		Subject: types.StringValue(pipeline.VCS.Commit.Subject),
		Body:    types.StringValue(pipeline.VCS.Commit.Body),
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vcsObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"provider_name":         types.StringType,
		"target_repository_url": types.StringType,
		"branch":                types.StringType,
		"review_id":             types.StringType,
		"review_url":            types.StringType,
		"revision":              types.StringType,
		"tag":                   types.StringType,
		"commit":                commitObj.Type(ctx),
	}, PipelineVCS{
		ProviderName:        types.StringValue(pipeline.VCS.ProviderName),
		TargetRepositoryURL: types.StringValue(pipeline.VCS.TargetRepositoryURL),
		Branch:              types.StringValue(pipeline.VCS.Branch),
		ReviewID:            types.StringValue(pipeline.VCS.ReviewID),
		ReviewURL:           types.StringValue(pipeline.VCS.ReviewURL),
		Revision:            types.StringValue(pipeline.VCS.Revision),
		Tag:                 types.StringValue(pipeline.VCS.Tag),
		Commit:              commitObj,
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.VCS = vcsObj

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PipelineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PipelineResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var pipeline Pipeline
	endpoint := fmt.Sprintf("/pipeline/%s", data.ID.ValueString())
	if err := r.client.Get(ctx, endpoint, &pipeline); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read pipeline, got error: %s", err))
		return
	}

	// Update the model with the response
	data.Number = types.Int64Value(pipeline.Number)
	data.State = types.StringValue(pipeline.State)
	data.CreatedAt = types.StringValue(pipeline.CreatedAt)
	data.UpdatedAt = types.StringValue(pipeline.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PipelineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Pipelines are immutable once created, so update is not supported
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Pipelines cannot be updated once created. To trigger a new pipeline, create a new resource.",
	)
}

func (r *PipelineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Pipelines cannot be deleted via API, but we can remove from state
	// In practice, you might want to cancel the pipeline if it's still running
	var data PipelineResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Try to cancel the pipeline if it's still running
	if data.State.ValueString() == "running" {
		endpoint := fmt.Sprintf("/pipeline/%s/cancel", data.ID.ValueString())
		if err := r.client.Post(ctx, endpoint, nil, nil); err != nil {
			// Don't fail deletion if cancel fails - just log it
			resp.Diagnostics.AddWarning("Pipeline Cancel Failed", fmt.Sprintf("Could not cancel running pipeline: %s", err))
		}
	}

	// Resource is automatically removed from state after successful Delete
}

func (r *PipelineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
