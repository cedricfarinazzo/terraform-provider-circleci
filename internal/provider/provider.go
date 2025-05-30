package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure CircleCIProvider satisfies various provider interfaces.
var _ provider.Provider = &CircleCIProvider{}

// CircleCIProvider defines the provider implementation.
type CircleCIProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// CircleCIProviderModel describes the provider data model.
type CircleCIProviderModel struct {
	ApiToken types.String `tfsdk:"api_token"`
	BaseURL  types.String `tfsdk:"base_url"`
}

// CircleCIClient holds the HTTP client and configuration for the CircleCI API
type CircleCIClient struct {
	ApiToken   string
	BaseURL    string
	HTTPClient *http.Client
}

func (p *CircleCIProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "circleci"
	resp.Version = p.version
}

func (p *CircleCIProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				MarkdownDescription: "CircleCI API token. Can also be set via the `CIRCLECI_TOKEN` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Base URL for CircleCI API. Defaults to https://circleci.com/api/v2",
				Optional:            true,
			},
		},
	}
}

func (p *CircleCIProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data CircleCIProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	var apiToken string
	var baseURL string

	if data.ApiToken.IsNull() {
		apiToken = os.Getenv("CIRCLECI_TOKEN")
	} else {
		apiToken = data.ApiToken.ValueString()
	}

	if data.BaseURL.IsNull() {
		baseURL = "https://circleci.com/api/v2"
	} else {
		baseURL = data.BaseURL.ValueString()
	}

	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Unable to find API token",
			"The CircleCI API token can be set in the provider configuration block api_token attribute or using the CIRCLECI_TOKEN environment variable. If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "circleci_api_token", apiToken)
	ctx = tflog.SetField(ctx, "circleci_base_url", baseURL)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "circleci_api_token")

	tflog.Debug(ctx, "Creating CircleCI client")

	// Create a CircleCI client and set it as the provider data
	client := &CircleCIClient{
		ApiToken:   apiToken,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}

	// Make the CircleCI client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured CircleCI client", map[string]any{"success": true})
}

func (p *CircleCIProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewContextResource,
		NewProjectResource,
		NewEnvironmentVariableResource,
		NewCheckoutKeyResource,
		NewWebhookResource,
		NewScheduleResource,
		NewPipelineResource,
		NewOIDCTokenResource,
		NewJobResource,
		NewPolicyResource,
		NewUserResource,
		NewUsageExportResource,
		NewRunnerResource,
		NewRunnerTokenResource,
	}
}

func (p *CircleCIProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewContextDataSource,
		NewProjectDataSource,
		NewInsightDataSource,
		NewOrganizationDataSource,
		NewWorkflowDataSource,
		NewWorkflowsDataSource,
		NewPoliciesDataSource,
		NewArtifactsDataSource,
		NewTestsDataSource,
		NewJobsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CircleCIProvider{
			version: version,
		}
	}
}
