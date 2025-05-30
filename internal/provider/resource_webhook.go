package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ resource.Resource = &WebhookResource{}
var _ resource.ResourceWithImportState = &WebhookResource{}

func NewWebhookResource() resource.Resource {
	return &WebhookResource{}
}

type WebhookResource struct {
	client *CircleCIClient
}

type WebhookResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	URL           types.String `tfsdk:"url"`
	Events        types.List   `tfsdk:"events"`
	SigningSecret types.String `tfsdk:"signing_secret"`
	VerifyTLS     types.Bool   `tfsdk:"verify_tls"`
	Scope         types.Object `tfsdk:"scope"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

type WebhookScope struct {
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

// CircleCI API models for webhooks
type Webhook struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	URL           string      `json:"url"`
	Events        []string    `json:"events"`
	SigningSecret string      `json:"signing_secret,omitempty"`
	VerifyTLS     bool        `json:"verify_tls"`
	Scope         ScopeObject `json:"scope"`
	CreatedAt     string      `json:"created_at"`
	UpdatedAt     string      `json:"updated_at"`
}

type ScopeObject struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type CreateWebhookRequest struct {
	Name          string      `json:"name"`
	URL           string      `json:"url"`
	Events        []string    `json:"events"`
	VerifyTLS     bool        `json:"verify_tls"`
	SigningSecret string      `json:"signing_secret,omitempty"`
	Scope         ScopeObject `json:"scope"`
}

func (r *WebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *WebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Webhook resource. Webhooks allow you to subscribe to events and receive HTTP notifications when those events occur.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the webhook.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the webhook.",
				Required:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The URL to which webhooks will be sent.",
				Required:            true,
			},
			"events": schema.ListAttribute{
				MarkdownDescription: "The events that will trigger this webhook. Valid events include: workflow-completed, job-completed, etc.",
				Required:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"signing_secret": schema.StringAttribute{
				MarkdownDescription: "The secret used to sign webhook payloads.",
				Optional:            true,
				Sensitive:           true,
			},
			"verify_tls": schema.BoolAttribute{
				MarkdownDescription: "Whether to verify TLS certificates when sending webhooks.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"scope": schema.SingleNestedAttribute{
				MarkdownDescription: "The scope of the webhook (organization or project).",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						MarkdownDescription: "The unique identifier of the scope.",
						Required:            true,
					},
					"type": schema.StringAttribute{
						MarkdownDescription: "The type of the scope ('project' or 'organization').",
						Required:            true,
					},
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time when the webhook was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time when the webhook was last updated.",
			},
		},
	}
}

func (r *WebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WebhookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract events list
	var events []string
	resp.Diagnostics.Append(data.Events.ElementsAs(ctx, &events, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract scope
	var scope WebhookScope
	resp.Diagnostics.Append(data.Scope.As(ctx, &scope, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := CreateWebhookRequest{
		Name:      data.Name.ValueString(),
		URL:       data.URL.ValueString(),
		Events:    events,
		VerifyTLS: data.VerifyTLS.ValueBool(),
		Scope: ScopeObject{
			ID:   scope.ID.ValueString(),
			Type: scope.Type.ValueString(),
		},
	}

	if !data.SigningSecret.IsNull() {
		createReq.SigningSecret = data.SigningSecret.ValueString()
	}

	var webhook Webhook
	if err := r.client.Post(ctx, "/webhook", createReq, &webhook); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create webhook, got error: %s", err))
		return
	}

	// Update the model with the response
	data.ID = types.StringValue(webhook.ID)
	data.CreatedAt = types.StringValue(webhook.CreatedAt)
	data.UpdatedAt = types.StringValue(webhook.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WebhookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var webhook Webhook
	if err := r.client.Get(ctx, fmt.Sprintf("/webhook/%s", data.ID.ValueString()), &webhook); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read webhook, got error: %s", err))
		return
	}

	// Update the model with the response
	data.Name = types.StringValue(webhook.Name)
	data.URL = types.StringValue(webhook.URL)
	data.VerifyTLS = types.BoolValue(webhook.VerifyTLS)
	data.CreatedAt = types.StringValue(webhook.CreatedAt)
	data.UpdatedAt = types.StringValue(webhook.UpdatedAt)

	// Convert events to types.List
	eventsList, diags := types.ListValueFrom(ctx, types.StringType, webhook.Events)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Events = eventsList

	// Convert scope to types.Object
	scopeObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"id":   types.StringType,
		"type": types.StringType,
	}, WebhookScope{
		ID:   types.StringValue(webhook.Scope.ID),
		Type: types.StringValue(webhook.Scope.Type),
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Scope = scopeObj

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WebhookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract events list
	var events []string
	resp.Diagnostics.Append(data.Events.ElementsAs(ctx, &events, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract scope
	var scope WebhookScope
	resp.Diagnostics.Append(data.Scope.As(ctx, &scope, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := CreateWebhookRequest{
		Name:      data.Name.ValueString(),
		URL:       data.URL.ValueString(),
		Events:    events,
		VerifyTLS: data.VerifyTLS.ValueBool(),
		Scope: ScopeObject{
			ID:   scope.ID.ValueString(),
			Type: scope.Type.ValueString(),
		},
	}

	if !data.SigningSecret.IsNull() {
		updateReq.SigningSecret = data.SigningSecret.ValueString()
	}

	var webhook Webhook
	if err := r.client.Put(ctx, fmt.Sprintf("/webhook/%s", data.ID.ValueString()), updateReq, &webhook); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update webhook, got error: %s", err))
		return
	}

	// Update the model with the response
	data.UpdatedAt = types.StringValue(webhook.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WebhookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, fmt.Sprintf("/webhook/%s", data.ID.ValueString())); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete webhook, got error: %s", err))
		return
	}
}

func (r *WebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
