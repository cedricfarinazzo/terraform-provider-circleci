package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &EnvironmentVariableResource{}
var _ resource.ResourceWithImportState = &EnvironmentVariableResource{}

func NewEnvironmentVariableResource() resource.Resource {
	return &EnvironmentVariableResource{}
}

type EnvironmentVariableResource struct {
	client *CircleCIClient
}

type EnvironmentVariableResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ContextID types.String `tfsdk:"context_id"`
	Name      types.String `tfsdk:"name"`
	Value     types.String `tfsdk:"value"`
}

// CircleCI API models for environment variables
type EnvironmentVariable struct {
	Variable  string `json:"variable"`
	Value     string `json:"value,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type CreateEnvironmentVariableRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (r *EnvironmentVariableResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_variable"
}

func (r *EnvironmentVariableResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Environment Variable resource. Environment variables can be set at the context level to be shared across projects.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the environment variable (format: context_id:variable_name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"context_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the context to add the environment variable to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the environment variable.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "The value of the environment variable.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *EnvironmentVariableResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EnvironmentVariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EnvironmentVariableResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := CreateEnvironmentVariableRequest{
		Name:  data.Name.ValueString(),
		Value: data.Value.ValueString(),
	}

	endpoint := fmt.Sprintf("/context/%s/environment-variable/%s",
		data.ContextID.ValueString(),
		data.Name.ValueString())

	var envVar EnvironmentVariable
	if err := r.client.Put(ctx, endpoint, createReq, &envVar); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create environment variable, got error: %s", err))
		return
	}

	// Set ID as combination of context_id and variable name
	data.ID = types.StringValue(fmt.Sprintf("%s:%s", data.ContextID.ValueString(), data.Name.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentVariableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EnvironmentVariableResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/context/%s/environment-variable/%s",
		data.ContextID.ValueString(),
		data.Name.ValueString())

	var envVar EnvironmentVariable
	if err := r.client.Get(ctx, endpoint, &envVar); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read environment variable, got error: %s", err))
		return
	}

	// Note: CircleCI API doesn't return the actual value for security reasons
	// We keep the value from the state/plan
	data.Name = types.StringValue(envVar.Variable)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentVariableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EnvironmentVariableResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := CreateEnvironmentVariableRequest{
		Name:  data.Name.ValueString(),
		Value: data.Value.ValueString(),
	}

	endpoint := fmt.Sprintf("/context/%s/environment-variable/%s",
		data.ContextID.ValueString(),
		data.Name.ValueString())

	var envVar EnvironmentVariable
	if err := r.client.Put(ctx, endpoint, updateReq, &envVar); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update environment variable, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentVariableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EnvironmentVariableResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/context/%s/environment-variable/%s",
		data.ContextID.ValueString(),
		data.Name.ValueString())

	if err := r.client.Delete(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete environment variable, got error: %s", err))
		return
	}
}

func (r *EnvironmentVariableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected import format: "context_id:variable_name"
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: context_id:variable_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("context_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
