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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ContextResource{}
var _ resource.ResourceWithImportState = &ContextResource{}

func NewContextResource() resource.Resource {
	return &ContextResource{}
}

// ContextResource defines the resource implementation.
type ContextResource struct {
	client *CircleCIClient
}

// ContextResourceModel describes the resource data model.
type ContextResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Owner types.Object `tfsdk:"owner"`
}

// ContextOwner represents the owner of a context
type ContextOwner struct {
	ID   types.String `tfsdk:"id"`
	Slug types.String `tfsdk:"slug"`
	Type types.String `tfsdk:"type"`
}

// CircleCI API models
type Context struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	Owner     Owner  `json:"owner"`
}

type Owner struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Type string `json:"type"`
}

type CreateContextRequest struct {
	Name  string `json:"name"`
	Owner Owner  `json:"owner"`
}

func (r *ContextResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_context"
}

func (r *ContextResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Context resource. Contexts provide a mechanism for securing and sharing environment variables across projects.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the context.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the context.",
				Required:            true,
			},
			"owner": schema.SingleNestedAttribute{
				MarkdownDescription: "The owner of the context (organization or user).",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						MarkdownDescription: "The unique identifier of the owner.",
						Optional:            true,
					},
					"slug": schema.StringAttribute{
						MarkdownDescription: "The slug of the owner (e.g., 'github', 'bitbucket').",
						Required:            true,
					},
					"type": schema.StringAttribute{
						MarkdownDescription: "The type of the owner ('organization' or 'account').",
						Required:            true,
					},
				},
			},
		},
	}
}

func (r *ContextResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (r *ContextResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ContextResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Extract owner information
	var owner ContextOwner
	resp.Diagnostics.Append(data.Owner.As(ctx, &owner, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create context request
	createReq := CreateContextRequest{
		Name: data.Name.ValueString(),
		Owner: Owner{
			ID:   owner.ID.ValueString(),
			Slug: owner.Slug.ValueString(),
			Type: owner.Type.ValueString(),
		},
	}

	var context Context
	if err := r.client.Post(ctx, "/context", createReq, &context); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create context, got error: %s", err))
		return
	}

	// Update the model with the response
	data.ID = types.StringValue(context.ID)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContextResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ContextResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var context Context
	if err := r.client.Get(ctx, fmt.Sprintf("/context/%s", data.ID.ValueString()), &context); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read context, got error: %s", err))
		return
	}

	// Update the model with the response
	data.Name = types.StringValue(context.Name)

	// Convert owner to types.Object
	ownerObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"id":   types.StringType,
			"slug": types.StringType,
			"type": types.StringType,
		},
		map[string]attr.Value{
			"id":   types.StringValue(context.Owner.ID),
			"slug": types.StringValue(context.Owner.Slug),
			"type": types.StringValue(context.Owner.Type),
		},
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Owner = ownerObj

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContextResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ContextResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Note: CircleCI API doesn't support updating context name or owner
	// This would require delete and recreate
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"CircleCI contexts cannot be updated. Changes to name or owner require destroying and recreating the context.",
	)
}

func (r *ContextResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ContextResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, fmt.Sprintf("/context/%s", data.ID.ValueString())); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete context, got error: %s", err))
		return
	}
}

func (r *ContextResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
