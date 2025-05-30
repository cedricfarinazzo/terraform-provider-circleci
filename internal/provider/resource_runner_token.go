package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RunnerTokenResource{}
var _ resource.ResourceWithImportState = &RunnerTokenResource{}

func NewRunnerTokenResource() resource.Resource {
	return &RunnerTokenResource{}
}

// RunnerTokenResource defines the resource implementation.
type RunnerTokenResource struct {
	client *CircleCIClient
}

// RunnerTokenResourceModel describes the resource data model.
type RunnerTokenResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ResourceClass types.String `tfsdk:"resource_class"`
	Nickname      types.String `tfsdk:"nickname"`
	Token         types.String `tfsdk:"token"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

func (r *RunnerTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runner_token"
}

func (r *RunnerTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a CircleCI runner authentication token.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the runner token.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_class": schema.StringAttribute{
				MarkdownDescription: "The resource class for which this token provides access.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nickname": schema.StringAttribute{
				MarkdownDescription: "A human-readable name for the token.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The authentication token value. This is only available when the token is first created.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The date and time the token was created.",
				Computed:            true,
			},
		},
	}
}

func (r *RunnerTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RunnerTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RunnerTokenResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API request body
	createReq := map[string]interface{}{
		"resource_class": data.ResourceClass.ValueString(),
		"nickname":       data.Nickname.ValueString(),
	}

	body, err := json.Marshal(createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating runner token",
			fmt.Sprintf("Unable to marshal request: %v", err),
		)
		return
	}

	// Make API request to create runner token
	url := fmt.Sprintf("%s/runner/token", r.client.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating request",
			fmt.Sprintf("Unable to create runner token request: %v", err),
		)
		return
	}

	httpReq.Header.Set("Circle-Token", r.client.ApiToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating runner token",
			fmt.Sprintf("Unable to create runner token: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Error creating runner token",
			fmt.Sprintf("API returned status %d when creating runner token", httpResp.StatusCode),
		)
		return
	}

	var apiResponse struct {
		ID            string `json:"id"`
		ResourceClass string `json:"resource_class"`
		Nickname      string `json:"nickname"`
		Token         string `json:"token"`
		CreatedAt     string `json:"created_at"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error parsing response",
			fmt.Sprintf("Unable to parse runner token creation response: %v", err),
		)
		return
	}

	// Update the model with response data
	data.ID = types.StringValue(apiResponse.ID)
	data.ResourceClass = types.StringValue(apiResponse.ResourceClass)
	data.Nickname = types.StringValue(apiResponse.Nickname)
	data.Token = types.StringValue(apiResponse.Token)
	data.CreatedAt = types.StringValue(apiResponse.CreatedAt)

	tflog.Trace(ctx, "created runner token resource", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RunnerTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RunnerTokenResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Make API request to get runner token
	url := fmt.Sprintf("%s/runner/token/%s", r.client.BaseURL, data.ID.ValueString())
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating request",
			fmt.Sprintf("Unable to create runner token request: %v", err),
		)
		return
	}

	httpReq.Header.Set("Circle-Token", r.client.ApiToken)
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading runner token",
			fmt.Sprintf("Unable to read runner token: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Error reading runner token",
			fmt.Sprintf("API returned status %d when reading runner token", httpResp.StatusCode),
		)
		return
	}

	var apiResponse struct {
		ID            string `json:"id"`
		ResourceClass string `json:"resource_class"`
		Nickname      string `json:"nickname"`
		CreatedAt     string `json:"created_at"`
		// Note: Token is not returned by the read API for security reasons
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error parsing response",
			fmt.Sprintf("Unable to parse runner token response: %v", err),
		)
		return
	}

	// Update the model with response data (keep existing token value)
	data.ID = types.StringValue(apiResponse.ID)
	data.ResourceClass = types.StringValue(apiResponse.ResourceClass)
	data.Nickname = types.StringValue(apiResponse.Nickname)
	data.CreatedAt = types.StringValue(apiResponse.CreatedAt)
	// Token remains as stored in state since it's not returned by read API

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RunnerTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Runner tokens are immutable - any changes require replacement
	resp.Diagnostics.AddError(
		"Runner token cannot be updated",
		"Runner tokens are immutable. Any changes require creating a new token.",
	)
}

func (r *RunnerTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RunnerTokenResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Make API request to delete runner token
	url := fmt.Sprintf("%s/runner/token/%s", r.client.BaseURL, data.ID.ValueString())
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating request",
			fmt.Sprintf("Unable to create runner token deletion request: %v", err),
		)
		return
	}

	httpReq.Header.Set("Circle-Token", r.client.ApiToken)

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting runner token",
			fmt.Sprintf("Unable to delete runner token: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting runner token",
			fmt.Sprintf("API returned status %d when deleting runner token", httpResp.StatusCode),
		)
		return
	}

	tflog.Trace(ctx, "deleted runner token resource", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
}

func (r *RunnerTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
