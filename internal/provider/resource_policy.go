package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &PolicyResource{}
var _ resource.ResourceWithImportState = &PolicyResource{}

func NewPolicyResource() resource.Resource {
	return &PolicyResource{}
}

type PolicyResource struct {
	client *CircleCIClient
}

type PolicyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Content     types.String `tfsdk:"content"`
	OrgID       types.String `tfsdk:"org_id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// CircleCI API models for policies
type PolicyAPI struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	OrgID       string `json:"org_id"`
	Enabled     bool   `json:"enabled"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type PolicyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Content     string `json:"content"`
	Enabled     bool   `json:"enabled"`
}

func (r *PolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (r *PolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Policy resource. This resource allows you to manage organization policies for controlling builds and workflows.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the policy.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The description of the policy.",
			},
			"content": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The policy content in OPA Rego format.",
			},
			"org_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The organization ID where the policy will be applied.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether the policy is enabled.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time the policy was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time the policy was last updated.",
			},
		},
	}
}

func (r *PolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgID.ValueString()
	policyRequest := PolicyRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Content:     data.Content.ValueString(),
		Enabled:     data.Enabled.ValueBool(),
	}

	endpoint := fmt.Sprintf("/v2/policy/%s", orgID)
	httpResp, err := r.client.MakeRequest(ctx, "POST", endpoint, policyRequest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create policy", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Failed to create policy",
			fmt.Sprintf("CircleCI API returned status %d", httpResp.StatusCode),
		)
		return
	}

	var policy PolicyAPI
	if err := json.NewDecoder(httpResp.Body).Decode(&policy); err != nil {
		resp.Diagnostics.AddError("Failed to decode response", err.Error())
		return
	}

	r.mapPolicyToModel(&policy, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgID.ValueString()
	policyID := data.ID.ValueString()

	endpoint := fmt.Sprintf("/v2/policy/%s/%s", orgID, policyID)
	httpResp, err := r.client.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get policy", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Failed to get policy",
			fmt.Sprintf("CircleCI API returned status %d", httpResp.StatusCode),
		)
		return
	}

	var policy PolicyAPI
	if err := json.NewDecoder(httpResp.Body).Decode(&policy); err != nil {
		resp.Diagnostics.AddError("Failed to decode response", err.Error())
		return
	}

	r.mapPolicyToModel(&policy, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgID.ValueString()
	policyID := data.ID.ValueString()

	policyRequest := PolicyRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Content:     data.Content.ValueString(),
		Enabled:     data.Enabled.ValueBool(),
	}

	endpoint := fmt.Sprintf("/v2/policy/%s/%s", orgID, policyID)
	httpResp, err := r.client.MakeRequest(ctx, "PUT", endpoint, policyRequest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update policy", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Failed to update policy",
			fmt.Sprintf("CircleCI API returned status %d", httpResp.StatusCode),
		)
		return
	}

	var policy PolicyAPI
	if err := json.NewDecoder(httpResp.Body).Decode(&policy); err != nil {
		resp.Diagnostics.AddError("Failed to decode response", err.Error())
		return
	}

	r.mapPolicyToModel(&policy, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgID.ValueString()
	policyID := data.ID.ValueString()

	endpoint := fmt.Sprintf("/v2/policy/%s/%s", orgID, policyID)
	httpResp, err := r.client.MakeRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete policy", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Failed to delete policy",
			fmt.Sprintf("CircleCI API returned status %d", httpResp.StatusCode),
		)
		return
	}
}

func (r *PolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "org_id:policy_id"
	// Example: "bb604b45-b6b0-4b81-ad80-796f15eddf87:550e8400-e29b-41d4-a716-446655440000"
	idParts := strings.SplitN(req.ID, ":", 2)
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Import ID must be in format 'org_id:policy_id', got: %q", req.ID),
		)
		return
	}

	orgID := idParts[0]
	policyID := idParts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), policyID)...)
}

func (r *PolicyResource) mapPolicyToModel(policy *PolicyAPI, data *PolicyResourceModel) {
	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.OrgID = types.StringValue(policy.OrgID)
	data.Content = types.StringValue(policy.Content)
	data.Enabled = types.BoolValue(policy.Enabled)
	data.CreatedAt = types.StringValue(policy.CreatedAt)
	data.UpdatedAt = types.StringValue(policy.UpdatedAt)

	if policy.Description != "" {
		data.Description = types.StringValue(policy.Description)
	} else {
		data.Description = types.StringNull()
	}
}
