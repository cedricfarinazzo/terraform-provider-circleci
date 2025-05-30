package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProjectDataSource{}

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

type ProjectDataSource struct {
	client *CircleCIClient
}

type ProjectDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Slug         types.String `tfsdk:"slug"`
	Name         types.String `tfsdk:"name"`
	Organization types.String `tfsdk:"organization"`
	VcsURL       types.String `tfsdk:"vcs_url"`
	VcsType      types.String `tfsdk:"vcs_type"`
}

func (d *ProjectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Project data source. Use this data source to get information about an existing project.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the project.",
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "The project slug in the form 'vcs-slug/org-name/repo-name'.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the project.",
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
				MarkdownDescription: "The version control system type.",
			},
		},
	}
}

func (d *ProjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := EscapeProjectSlug(data.Slug.ValueString())
	endpoint := fmt.Sprintf("/project/%s", slug)

	var project Project
	if err := d.client.Get(ctx, endpoint, &project); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read project, got error: %s", err))
		return
	}

	// Update the model with the response
	data.ID = types.StringValue(project.ID)
	data.Name = types.StringValue(project.Name)
	data.Organization = types.StringValue(project.Organization)
	data.VcsURL = types.StringValue(project.VcsInfo.VcsURL)
	data.VcsType = types.StringValue(project.VcsInfo.Provider)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
