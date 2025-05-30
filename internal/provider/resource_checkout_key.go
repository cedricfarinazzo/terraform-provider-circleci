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

var _ resource.Resource = &CheckoutKeyResource{}
var _ resource.ResourceWithImportState = &CheckoutKeyResource{}

func NewCheckoutKeyResource() resource.Resource {
	return &CheckoutKeyResource{}
}

type CheckoutKeyResource struct {
	client *CircleCIClient
}

type CheckoutKeyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectSlug types.String `tfsdk:"project_slug"`
	Type        types.String `tfsdk:"type"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	PublicKey   types.String `tfsdk:"public_key"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

// CircleCI API models for checkout keys
type CheckoutKey struct {
	PublicKey   string `json:"public_key"`
	Type        string `json:"type"`
	Fingerprint string `json:"fingerprint"`
	Preferred   bool   `json:"preferred"`
	CreatedAt   string `json:"created_at"`
}

type CreateCheckoutKeyRequest struct {
	Type string `json:"type"`
}

func (r *CheckoutKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_checkout_key"
}

func (r *CheckoutKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Checkout Key resource. Checkout keys are used to access your repository during builds.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the checkout key (format: project_slug:fingerprint).",
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
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of checkout key. Valid values are 'user-key' and 'deploy-key'.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"fingerprint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The SSH fingerprint of the checkout key.",
			},
			"public_key": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The public SSH key.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time when the checkout key was created.",
			},
		},
	}
}

func (r *CheckoutKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CheckoutKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CheckoutKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := CreateCheckoutKeyRequest{
		Type: data.Type.ValueString(),
	}

	slug := EscapeProjectSlug(data.ProjectSlug.ValueString())
	endpoint := fmt.Sprintf("/project/%s/checkout-key", slug)

	var checkoutKey CheckoutKey
	if err := r.client.Post(ctx, endpoint, createReq, &checkoutKey); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create checkout key, got error: %s", err))
		return
	}

	// Update the model with the response
	data.ID = types.StringValue(fmt.Sprintf("%s:%s", data.ProjectSlug.ValueString(), checkoutKey.Fingerprint))
	data.Fingerprint = types.StringValue(checkoutKey.Fingerprint)
	data.PublicKey = types.StringValue(checkoutKey.PublicKey)
	data.CreatedAt = types.StringValue(checkoutKey.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CheckoutKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CheckoutKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := EscapeProjectSlug(data.ProjectSlug.ValueString())
	endpoint := fmt.Sprintf("/project/%s/checkout-key/%s", slug, data.Fingerprint.ValueString())

	var checkoutKey CheckoutKey
	if err := r.client.Get(ctx, endpoint, &checkoutKey); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read checkout key, got error: %s", err))
		return
	}

	// Update the model with the response
	data.Type = types.StringValue(checkoutKey.Type)
	data.PublicKey = types.StringValue(checkoutKey.PublicKey)
	data.CreatedAt = types.StringValue(checkoutKey.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CheckoutKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Checkout keys cannot be updated - they need to be recreated
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"CircleCI checkout keys cannot be updated. Changes require destroying and recreating the key.",
	)
}

func (r *CheckoutKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CheckoutKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := EscapeProjectSlug(data.ProjectSlug.ValueString())
	endpoint := fmt.Sprintf("/project/%s/checkout-key/%s", slug, data.Fingerprint.ValueString())

	if err := r.client.Delete(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete checkout key, got error: %s", err))
		return
	}
}

func (r *CheckoutKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected import format: "project_slug:fingerprint"
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: project_slug:fingerprint. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_slug"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("fingerprint"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
