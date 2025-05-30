package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ContextDataSource{}

func NewContextDataSource() datasource.DataSource {
	return &ContextDataSource{}
}

type ContextDataSource struct {
	client *CircleCIClient
}

type ContextDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Owner types.Object `tfsdk:"owner"`
}

func (d *ContextDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_context"
}

func (d *ContextDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Context data source. Use this data source to get information about an existing context.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the context. Either 'id' or 'name' must be specified.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the context. Either 'id' or 'name' must be specified.",
				Optional:            true,
				Computed:            true,
			},
			"owner": schema.SingleNestedAttribute{
				MarkdownDescription: "The owner of the context.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						MarkdownDescription: "The unique identifier of the owner.",
						Computed:            true,
					},
					"slug": schema.StringAttribute{
						MarkdownDescription: "The slug of the owner.",
						Computed:            true,
					},
					"type": schema.StringAttribute{
						MarkdownDescription: "The type of the owner.",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (d *ContextDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*CircleCIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *CircleCIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ContextDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ContextDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var context Context

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		// Read by ID
		if err := d.client.Get(ctx, fmt.Sprintf("/context/%s", data.ID.ValueString()), &context); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read context, got error: %s", err))
			return
		}
	} else if !data.Name.IsNull() && !data.Name.IsUnknown() {
		// Read by name - need to list all contexts and find by name
		var contexts struct {
			Items []Context `json:"items"`
		}
		if err := d.client.Get(ctx, "/context", &contexts); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list contexts, got error: %s", err))
			return
		}

		var found bool
		for _, ctx := range contexts.Items {
			if ctx.Name == data.Name.ValueString() {
				context = ctx
				found = true
				break
			}
		}

		if !found {
			resp.Diagnostics.AddError("Context Not Found", fmt.Sprintf("Context with name '%s' not found", data.Name.ValueString()))
			return
		}
	} else {
		resp.Diagnostics.AddError("Missing Required Attribute", "Either 'id' or 'name' must be specified")
		return
	}

	// Update the model with the response
	data.ID = types.StringValue(context.ID)
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
