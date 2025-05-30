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

var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	client *CircleCIClient
}

type UserResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Login    types.String `tfsdk:"login"`
	Name     types.String `tfsdk:"name"`
	Email    types.String `tfsdk:"email"`
	OrgID    types.String `tfsdk:"org_id"`
	Role     types.String `tfsdk:"role"`
	Avatar   types.String `tfsdk:"avatar_url"`
	JoinedAt types.String `tfsdk:"joined_at"`
}

// CircleCI API models for users
type UserAPI struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar_url"`
	JoinedAt string `json:"joined_at"`
}

type UserInviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UserRoleRequest struct {
	Role string `json:"role"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI User resource. This resource allows you to manage team members in your organization.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"login": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The user's login name.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The user's full name.",
			},
			"email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The user's email address.",
			},
			"org_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The organization ID where the user belongs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The user's role in the organization (admin, member, viewer).",
			},
			"avatar_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The user's avatar URL.",
			},
			"joined_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time the user joined the organization.",
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgID.ValueString()
	inviteRequest := UserInviteRequest{
		Email: data.Email.ValueString(),
		Role:  data.Role.ValueString(),
	}

	// Invite user to organization
	endpoint := fmt.Sprintf("/v2/organization/%s/invite", orgID)
	httpResp, err := r.client.MakeRequest(ctx, "POST", endpoint, inviteRequest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to invite user", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Failed to invite user",
			fmt.Sprintf("CircleCI API returned status %d", httpResp.StatusCode),
		)
		return
	}

	var user UserAPI
	if err := json.NewDecoder(httpResp.Body).Decode(&user); err != nil {
		resp.Diagnostics.AddError("Failed to decode response", err.Error())
		return
	}

	r.mapUserToModel(&user, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgID.ValueString()
	userID := data.ID.ValueString()

	endpoint := fmt.Sprintf("/v2/organization/%s/user/%s", orgID, userID)
	httpResp, err := r.client.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get user", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Failed to get user",
			fmt.Sprintf("CircleCI API returned status %d", httpResp.StatusCode),
		)
		return
	}

	var user UserAPI
	if err := json.NewDecoder(httpResp.Body).Decode(&user); err != nil {
		resp.Diagnostics.AddError("Failed to decode response", err.Error())
		return
	}

	r.mapUserToModel(&user, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgID.ValueString()
	userID := data.ID.ValueString()

	roleRequest := UserRoleRequest{
		Role: data.Role.ValueString(),
	}

	endpoint := fmt.Sprintf("/v2/organization/%s/user/%s/role", orgID, userID)
	httpResp, err := r.client.MakeRequest(ctx, "PUT", endpoint, roleRequest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update user role", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Failed to update user role",
			fmt.Sprintf("CircleCI API returned status %d", httpResp.StatusCode),
		)
		return
	}

	// Read updated user data
	r.Read(ctx, resource.ReadRequest{State: req.State}, &resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics})
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgID.ValueString()
	userID := data.ID.ValueString()

	endpoint := fmt.Sprintf("/v2/organization/%s/user/%s", orgID, userID)
	httpResp, err := r.client.MakeRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to remove user", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Failed to remove user",
			fmt.Sprintf("CircleCI API returned status %d", httpResp.StatusCode),
		)
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "org_id:user_id"
	// Example: "bb604b45-b6b0-4b81-ad80-796f15eddf87:550e8400-e29b-41d4-a716-446655440000"
	idParts := strings.SplitN(req.ID, ":", 2)
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Import ID must be in format 'org_id:user_id', got: %q", req.ID),
		)
		return
	}

	orgID := idParts[0]
	userID := idParts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), userID)...)
}

func (r *UserResource) mapUserToModel(user *UserAPI, data *UserResourceModel) {
	data.ID = types.StringValue(user.ID)
	data.Login = types.StringValue(user.Login)
	data.Name = types.StringValue(user.Name)
	data.Email = types.StringValue(user.Email)
	data.Avatar = types.StringValue(user.Avatar)
	data.JoinedAt = types.StringValue(user.JoinedAt)
}
