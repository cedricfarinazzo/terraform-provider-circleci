package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &OrganizationDataSource{}

func NewOrganizationDataSource() datasource.DataSource {
	return &OrganizationDataSource{}
}

type OrganizationDataSource struct {
	client *CircleCIClient
}

type OrganizationDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Slug      types.String `tfsdk:"slug"`
	VcsType   types.String `tfsdk:"vcs_type"`
	AvatarURL types.String `tfsdk:"avatar_url"`
	CreatedAt types.String `tfsdk:"created_at"`
}

// CircleCI API models for organizations
type Organization struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	VcsType   string `json:"vcs_type"`
	AvatarURL string `json:"avatar_url"`
	CreatedAt string `json:"created_at"`
}

func (d *OrganizationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *OrganizationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Organization data source. Use this data source to get information about an organization.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the organization. Either 'id' or 'name' must be specified.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the organization. Either 'id' or 'name' must be specified.",
				Optional:            true,
				Computed:            true,
			},
			"slug": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The slug of the organization.",
			},
			"vcs_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The version control system type (e.g., 'github', 'bitbucket').",
			},
			"avatar_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL of the organization's avatar image.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time when the organization was created.",
			},
		},
	}
}

func (d *OrganizationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var org Organization

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		// Read by ID
		if err := d.client.Get(ctx, fmt.Sprintf("/organization/%s", data.ID.ValueString()), &org); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read organization, got error: %s", err))
			return
		}
	} else if !data.Name.IsNull() && !data.Name.IsUnknown() {
		// Read by name - need to list all organizations and find by name
		var orgs struct {
			Items []Organization `json:"items"`
		}
		if err := d.client.Get(ctx, "/organization", &orgs); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list organizations, got error: %s", err))
			return
		}

		var found bool
		for _, organization := range orgs.Items {
			if organization.Name == data.Name.ValueString() {
				org = organization
				found = true
				break
			}
		}

		if !found {
			resp.Diagnostics.AddError("Organization Not Found", fmt.Sprintf("Organization with name '%s' not found", data.Name.ValueString()))
			return
		}
	} else {
		resp.Diagnostics.AddError("Missing Required Attribute", "Either 'id' or 'name' must be specified")
		return
	}

	// Update the model with the response
	data.ID = types.StringValue(org.ID)
	data.Name = types.StringValue(org.Name)
	data.Slug = types.StringValue(org.Slug)
	data.VcsType = types.StringValue(org.VcsType)
	data.AvatarURL = types.StringValue(org.AvatarURL)
	data.CreatedAt = types.StringValue(org.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
