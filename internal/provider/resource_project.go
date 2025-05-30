package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

type ProjectResource struct {
	client *CircleCIClient
}

type ProjectResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Slug         types.String `tfsdk:"slug"`
	Name         types.String `tfsdk:"name"`
	Organization types.String `tfsdk:"organization"`
	VcsURL       types.String `tfsdk:"vcs_url"`
	VcsType      types.String `tfsdk:"vcs_type"`
}

// CircleCI API models for projects
type Project struct {
	ID           string  `json:"id"`
	Slug         string  `json:"slug"`
	Name         string  `json:"name"`
	Organization string  `json:"organization_name"`
	VcsInfo      VcsInfo `json:"vcs_info"`
}

type VcsInfo struct {
	VcsURL        string `json:"vcs_url"`
	Provider      string `json:"provider"`
	DefaultBranch string `json:"default_branch"`
}

func (r *ProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Project resource. Projects contain the build configuration and history for a repository.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "The project slug in the form 'vcs-slug/org-name/repo-name' (e.g., 'gh/circleci/circleci-docs').",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the project (repository name).",
			},
			"organization": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the organization that owns the project.",
			},
			"vcs_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL of the repository.",
			},
			"vcs_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The version control system type (e.g., 'github', 'bitbucket').",
			},
		},
	}
}

func (r *ProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// For CircleCI, "creating" a project typically means following it
	// The project must already exist in the VCS (GitHub, Bitbucket, etc.)
	slug := EscapeProjectSlug(data.Slug.ValueString())
	endpoint := fmt.Sprintf("/project/%s/follow", slug)

	// Follow the project
	if err := r.client.Post(ctx, endpoint, nil, nil); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to follow project, got error: %s", err))
		return
	}

	// Read the project details
	r.readProject(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readProject(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) readProject(ctx context.Context, data *ProjectResourceModel, diagnostics *diag.Diagnostics) {
	slug := EscapeProjectSlug(data.Slug.ValueString())
	endpoint := fmt.Sprintf("/project/%s", slug)

	var project Project
	if err := r.client.Get(ctx, endpoint, &project); err != nil {
		diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read project, got error: %s", err))
		return
	}

	data.ID = types.StringValue(project.ID)
	data.Name = types.StringValue(project.Name)
	data.Organization = types.StringValue(project.Organization)
	data.VcsURL = types.StringValue(project.VcsInfo.VcsURL)
	data.VcsType = types.StringValue(project.VcsInfo.Provider)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// CircleCI projects are mostly read-only - the primary data comes from the VCS
	// Most updates would be to project settings, which would be separate resources
	resp.Diagnostics.AddWarning(
		"Project Update",
		"CircleCI projects are primarily read-only. Most project settings are managed through separate resources.",
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// For CircleCI, "deleting" a project means unfollowing it
	slug := EscapeProjectSlug(data.Slug.ValueString())
	endpoint := fmt.Sprintf("/project/%s/unfollow", slug)

	if err := r.client.Post(ctx, endpoint, nil, nil); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to unfollow project, got error: %s", err))
		return
	}
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using the project slug
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), req.ID)...)
}
