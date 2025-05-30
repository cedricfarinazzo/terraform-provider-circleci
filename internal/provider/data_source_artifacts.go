package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ArtifactsDataSource{}

func NewArtifactsDataSource() datasource.DataSource {
	return &ArtifactsDataSource{}
}

// ArtifactsDataSource defines the data source implementation.
type ArtifactsDataSource struct {
	client *CircleCIClient
}

// ArtifactsDataSourceModel describes the data source data model.
type ArtifactsDataSourceModel struct {
	ProjectSlug types.String              `tfsdk:"project_slug"`
	JobNumber   types.Int64               `tfsdk:"job_number"`
	Artifacts   []ArtifactDataSourceModel `tfsdk:"artifacts"`
}

type ArtifactDataSourceModel struct {
	Path       types.String `tfsdk:"path"`
	NodeIndex  types.Int64  `tfsdk:"node_index"`
	URL        types.String `tfsdk:"url"`
	PrettyPath types.String `tfsdk:"pretty_path"`
}

func (d *ArtifactsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_artifacts"
}

func (d *ArtifactsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves artifacts from a CircleCI job.",

		Attributes: map[string]schema.Attribute{
			"project_slug": schema.StringAttribute{
				MarkdownDescription: "Project slug in the form `vcs-slug/org-name/repo-name`.",
				Required:            true,
			},
			"job_number": schema.Int64Attribute{
				MarkdownDescription: "The job number for which to retrieve artifacts.",
				Required:            true,
			},
			"artifacts": schema.ListNestedAttribute{
				MarkdownDescription: "List of artifacts from the job.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							MarkdownDescription: "The path to the artifact file.",
							Computed:            true,
						},
						"node_index": schema.Int64Attribute{
							MarkdownDescription: "The node index for the artifact.",
							Computed:            true,
						},
						"url": schema.StringAttribute{
							MarkdownDescription: "The download URL for the artifact.",
							Computed:            true,
						},
						"pretty_path": schema.StringAttribute{
							MarkdownDescription: "The pretty-formatted path to the artifact.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *ArtifactsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (d *ArtifactsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ArtifactsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Make API request to get artifacts
	url := fmt.Sprintf("%s/project/%s/%d/artifacts", d.client.BaseURL, data.ProjectSlug.ValueString(), data.JobNumber.ValueInt64())

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating request",
			fmt.Sprintf("Unable to create request for artifacts: %v", err),
		)
		return
	}

	httpReq.Header.Set("Circle-Token", d.client.ApiToken)
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making request",
			fmt.Sprintf("Unable to get artifacts: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Error from API",
			fmt.Sprintf("API returned status %d when getting artifacts", httpResp.StatusCode),
		)
		return
	}

	var apiResponse struct {
		Items []struct {
			Path       string `json:"path"`
			NodeIndex  int64  `json:"node_index"`
			URL        string `json:"url"`
			PrettyPath string `json:"pretty_path"`
		} `json:"items"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error parsing response",
			fmt.Sprintf("Unable to parse artifacts response: %v", err),
		)
		return
	}

	// Convert API response to model
	data.Artifacts = make([]ArtifactDataSourceModel, len(apiResponse.Items))
	for i, artifact := range apiResponse.Items {
		data.Artifacts[i] = ArtifactDataSourceModel{
			Path:       types.StringValue(artifact.Path),
			NodeIndex:  types.Int64Value(artifact.NodeIndex),
			URL:        types.StringValue(artifact.URL),
			PrettyPath: types.StringValue(artifact.PrettyPath),
		}
	}

	tflog.Trace(ctx, "read artifacts data source", map[string]interface{}{
		"project_slug": data.ProjectSlug.ValueString(),
		"job_number":   data.JobNumber.ValueInt64(),
		"count":        len(data.Artifacts),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
