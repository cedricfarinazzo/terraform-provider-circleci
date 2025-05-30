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

var _ resource.Resource = &OIDCTokenResource{}
var _ resource.ResourceWithImportState = &OIDCTokenResource{}

func NewOIDCTokenResource() resource.Resource {
	return &OIDCTokenResource{}
}

type OIDCTokenResource struct {
	client *CircleCIClient
}

type OIDCTokenResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	OrgID       types.String `tfsdk:"org_id"`
	Audience    types.String `tfsdk:"audience"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// CircleCI API models for OIDC tokens
type OIDCToken struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	OrgID       string `json:"org_id"`
	Audience    string `json:"audience"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateOIDCTokenRequest struct {
	Name        string `json:"name"`
	Audience    string `json:"audience"`
	Description string `json:"description,omitempty"`
}

func (r *OIDCTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oidc_token"
}

func (r *OIDCTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI OIDC Token resource. OIDC tokens allow you to use OpenID Connect to authenticate with external services without storing long-lived credentials.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the OIDC token.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the OIDC token.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"org_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID that owns this OIDC token.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"audience": schema.StringAttribute{
				MarkdownDescription: "The audience for the OIDC token. This should be the identifier of the external service that will consume the token.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the OIDC token.",
				Optional:            true,
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time when the OIDC token was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time when the OIDC token was last updated.",
			},
		},
	}
}

func (r *OIDCTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OIDCTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OIDCTokenResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := CreateOIDCTokenRequest{
		Name:        data.Name.ValueString(),
		Audience:    data.Audience.ValueString(),
		Description: data.Description.ValueString(),
	}

	endpoint := fmt.Sprintf("/organization/%s/oidc-token", data.OrgID.ValueString())

	var token OIDCToken
	if err := r.client.Post(ctx, endpoint, createReq, &token); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create OIDC token, got error: %s", err))
		return
	}

	// Update the model with the response
	data.ID = types.StringValue(token.ID)
	data.CreatedAt = types.StringValue(token.CreatedAt)
	data.UpdatedAt = types.StringValue(token.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OIDCTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OIDCTokenResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/organization/%s/oidc-token/%s", data.OrgID.ValueString(), data.ID.ValueString())

	var token OIDCToken
	if err := r.client.Get(ctx, endpoint, &token); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read OIDC token, got error: %s", err))
		return
	}

	// Update the model with the response
	data.Name = types.StringValue(token.Name)
	data.Audience = types.StringValue(token.Audience)
	data.Description = types.StringValue(token.Description)
	data.CreatedAt = types.StringValue(token.CreatedAt)
	data.UpdatedAt = types.StringValue(token.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OIDCTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OIDCTokenResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := CreateOIDCTokenRequest{
		Name:        data.Name.ValueString(),
		Audience:    data.Audience.ValueString(),
		Description: data.Description.ValueString(),
	}

	endpoint := fmt.Sprintf("/organization/%s/oidc-token/%s", data.OrgID.ValueString(), data.ID.ValueString())

	var token OIDCToken
	if err := r.client.Put(ctx, endpoint, updateReq, &token); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update OIDC token, got error: %s", err))
		return
	}

	// Update the model with the response
	data.UpdatedAt = types.StringValue(token.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OIDCTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OIDCTokenResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/organization/%s/oidc-token/%s", data.OrgID.ValueString(), data.ID.ValueString())

	if err := r.client.Delete(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete OIDC token, got error: %s", err))
		return
	}
}

func (r *OIDCTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected import format: "org_id:token_id"
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id:token_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
