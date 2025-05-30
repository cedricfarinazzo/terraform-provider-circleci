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
var _ datasource.DataSource = &TestsDataSource{}

func NewTestsDataSource() datasource.DataSource {
	return &TestsDataSource{}
}

// TestsDataSource defines the data source implementation.
type TestsDataSource struct {
	client *CircleCIClient
}

// TestsDataSourceModel describes the data source data model.
type TestsDataSourceModel struct {
	ProjectSlug types.String          `tfsdk:"project_slug"`
	JobNumber   types.Int64           `tfsdk:"job_number"`
	Tests       []TestDataSourceModel `tfsdk:"tests"`
}

type TestDataSourceModel struct {
	Message   types.String  `tfsdk:"message"`
	Source    types.String  `tfsdk:"source"`
	RunTime   types.Float64 `tfsdk:"run_time"`
	File      types.String  `tfsdk:"file"`
	Result    types.String  `tfsdk:"result"`
	Name      types.String  `tfsdk:"name"`
	Classname types.String  `tfsdk:"classname"`
}

func (d *TestsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tests"
}

func (d *TestsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves test results from a CircleCI job.",

		Attributes: map[string]schema.Attribute{
			"project_slug": schema.StringAttribute{
				MarkdownDescription: "Project slug in the form `vcs-slug/org-name/repo-name`.",
				Required:            true,
			},
			"job_number": schema.Int64Attribute{
				MarkdownDescription: "The job number for which to retrieve test results.",
				Required:            true,
			},
			"tests": schema.ListNestedAttribute{
				MarkdownDescription: "List of test results from the job.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"message": schema.StringAttribute{
							MarkdownDescription: "The test result message.",
							Computed:            true,
						},
						"source": schema.StringAttribute{
							MarkdownDescription: "The test source.",
							Computed:            true,
						},
						"run_time": schema.Float64Attribute{
							MarkdownDescription: "The test execution time in seconds.",
							Computed:            true,
						},
						"file": schema.StringAttribute{
							MarkdownDescription: "The test file path.",
							Computed:            true,
						},
						"result": schema.StringAttribute{
							MarkdownDescription: "The test result (success, failure, skipped).",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The test name.",
							Computed:            true,
						},
						"classname": schema.StringAttribute{
							MarkdownDescription: "The test class name.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *TestsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TestsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TestsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Make API request to get test results
	url := fmt.Sprintf("%s/project/%s/%d/tests", d.client.BaseURL, data.ProjectSlug.ValueString(), data.JobNumber.ValueInt64())

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating request",
			fmt.Sprintf("Unable to create request for tests: %v", err),
		)
		return
	}

	httpReq.Header.Set("Circle-Token", d.client.ApiToken)
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making request",
			fmt.Sprintf("Unable to get tests: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Error from API",
			fmt.Sprintf("API returned status %d when getting tests", httpResp.StatusCode),
		)
		return
	}

	var apiResponse struct {
		Items []struct {
			Message   string  `json:"message"`
			Source    string  `json:"source"`
			RunTime   float64 `json:"run_time"`
			File      string  `json:"file"`
			Result    string  `json:"result"`
			Name      string  `json:"name"`
			Classname string  `json:"classname"`
		} `json:"items"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error parsing response",
			fmt.Sprintf("Unable to parse tests response: %v", err),
		)
		return
	}

	// Convert API response to model
	data.Tests = make([]TestDataSourceModel, len(apiResponse.Items))
	for i, test := range apiResponse.Items {
		data.Tests[i] = TestDataSourceModel{
			Message:   types.StringValue(test.Message),
			Source:    types.StringValue(test.Source),
			RunTime:   types.Float64Value(test.RunTime),
			File:      types.StringValue(test.File),
			Result:    types.StringValue(test.Result),
			Name:      types.StringValue(test.Name),
			Classname: types.StringValue(test.Classname),
		}
	}

	tflog.Trace(ctx, "read tests data source", map[string]interface{}{
		"project_slug": data.ProjectSlug.ValueString(),
		"job_number":   data.JobNumber.ValueInt64(),
		"count":        len(data.Tests),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
