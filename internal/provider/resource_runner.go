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
var _ resource.Resource = &RunnerResource{}
var _ resource.ResourceWithImportState = &RunnerResource{}

func NewRunnerResource() resource.Resource {
	return &RunnerResource{}
}

// RunnerResource defines the resource implementation.
type RunnerResource struct {
	client *CircleCIClient
}

// RunnerResourceModel describes the resource data model.
type RunnerResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	ResourceClass  types.String `tfsdk:"resource_class"`
	Platform       types.String `tfsdk:"platform"`
	IP             types.String `tfsdk:"ip"`
	Hostname       types.String `tfsdk:"hostname"`
	Version        types.String `tfsdk:"version"`
	FirstConnected types.String `tfsdk:"first_connected"`
	LastConnected  types.String `tfsdk:"last_connected"`
	LastUsed       types.String `tfsdk:"last_used"`
	State          types.String `tfsdk:"state"`
}

func (r *RunnerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runner"
}

func (r *RunnerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a CircleCI self-hosted runner.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the runner.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the runner.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the runner.",
				Optional:            true,
			},
			"resource_class": schema.StringAttribute{
				MarkdownDescription: "The resource class for the runner.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"platform": schema.StringAttribute{
				MarkdownDescription: "The platform of the runner (linux, windows, darwin).",
				Computed:            true,
			},
			"ip": schema.StringAttribute{
				MarkdownDescription: "The IP address of the runner.",
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "The hostname of the runner.",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The version of the runner agent.",
				Computed:            true,
			},
			"first_connected": schema.StringAttribute{
				MarkdownDescription: "The date and time the runner first connected.",
				Computed:            true,
			},
			"last_connected": schema.StringAttribute{
				MarkdownDescription: "The date and time the runner last connected.",
				Computed:            true,
			},
			"last_used": schema.StringAttribute{
				MarkdownDescription: "The date and time the runner was last used.",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The current state of the runner.",
				Computed:            true,
			},
		},
	}
}

func (r *RunnerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RunnerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RunnerResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API request body
	createReq := map[string]interface{}{
		"name":           data.Name.ValueString(),
		"resource_class": data.ResourceClass.ValueString(),
	}

	if !data.Description.IsNull() {
		createReq["description"] = data.Description.ValueString()
	}

	body, err := json.Marshal(createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating runner",
			fmt.Sprintf("Unable to marshal request: %v", err),
		)
		return
	}

	// Make API request to create runner
	url := fmt.Sprintf("%s/runner", r.client.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating request",
			fmt.Sprintf("Unable to create runner request: %v", err),
		)
		return
	}

	httpReq.Header.Set("Circle-Token", r.client.ApiToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating runner",
			fmt.Sprintf("Unable to create runner: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Error creating runner",
			fmt.Sprintf("API returned status %d when creating runner", httpResp.StatusCode),
		)
		return
	}

	var apiResponse struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Description    string `json:"description"`
		ResourceClass  string `json:"resource_class"`
		Platform       string `json:"platform"`
		IP             string `json:"ip"`
		Hostname       string `json:"hostname"`
		Version        string `json:"version"`
		FirstConnected string `json:"first_connected"`
		LastConnected  string `json:"last_connected"`
		LastUsed       string `json:"last_used"`
		State          string `json:"state"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error parsing response",
			fmt.Sprintf("Unable to parse runner creation response: %v", err),
		)
		return
	}

	// Update the model with response data
	data.ID = types.StringValue(apiResponse.ID)
	data.Name = types.StringValue(apiResponse.Name)
	data.Description = types.StringValue(apiResponse.Description)
	data.ResourceClass = types.StringValue(apiResponse.ResourceClass)
	data.Platform = types.StringValue(apiResponse.Platform)
	data.IP = types.StringValue(apiResponse.IP)
	data.Hostname = types.StringValue(apiResponse.Hostname)
	data.Version = types.StringValue(apiResponse.Version)
	data.FirstConnected = types.StringValue(apiResponse.FirstConnected)
	data.LastConnected = types.StringValue(apiResponse.LastConnected)
	data.LastUsed = types.StringValue(apiResponse.LastUsed)
	data.State = types.StringValue(apiResponse.State)

	tflog.Trace(ctx, "created runner resource", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RunnerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RunnerResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Make API request to get runner
	url := fmt.Sprintf("%s/runner/%s", r.client.BaseURL, data.ID.ValueString())
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating request",
			fmt.Sprintf("Unable to create runner request: %v", err),
		)
		return
	}

	httpReq.Header.Set("Circle-Token", r.client.ApiToken)
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading runner",
			fmt.Sprintf("Unable to read runner: %v", err),
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
			"Error reading runner",
			fmt.Sprintf("API returned status %d when reading runner", httpResp.StatusCode),
		)
		return
	}

	var apiResponse struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Description    string `json:"description"`
		ResourceClass  string `json:"resource_class"`
		Platform       string `json:"platform"`
		IP             string `json:"ip"`
		Hostname       string `json:"hostname"`
		Version        string `json:"version"`
		FirstConnected string `json:"first_connected"`
		LastConnected  string `json:"last_connected"`
		LastUsed       string `json:"last_used"`
		State          string `json:"state"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error parsing response",
			fmt.Sprintf("Unable to parse runner response: %v", err),
		)
		return
	}

	// Update the model with response data
	data.ID = types.StringValue(apiResponse.ID)
	data.Name = types.StringValue(apiResponse.Name)
	data.Description = types.StringValue(apiResponse.Description)
	data.ResourceClass = types.StringValue(apiResponse.ResourceClass)
	data.Platform = types.StringValue(apiResponse.Platform)
	data.IP = types.StringValue(apiResponse.IP)
	data.Hostname = types.StringValue(apiResponse.Hostname)
	data.Version = types.StringValue(apiResponse.Version)
	data.FirstConnected = types.StringValue(apiResponse.FirstConnected)
	data.LastConnected = types.StringValue(apiResponse.LastConnected)
	data.LastUsed = types.StringValue(apiResponse.LastUsed)
	data.State = types.StringValue(apiResponse.State)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RunnerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RunnerResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API request body
	updateReq := map[string]interface{}{
		"name": data.Name.ValueString(),
	}

	if !data.Description.IsNull() {
		updateReq["description"] = data.Description.ValueString()
	}

	body, err := json.Marshal(updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating runner",
			fmt.Sprintf("Unable to marshal request: %v", err),
		)
		return
	}

	// Make API request to update runner
	url := fmt.Sprintf("%s/runner/%s", r.client.BaseURL, data.ID.ValueString())
	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", url, strings.NewReader(string(body)))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating request",
			fmt.Sprintf("Unable to create runner update request: %v", err),
		)
		return
	}

	httpReq.Header.Set("Circle-Token", r.client.ApiToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating runner",
			fmt.Sprintf("Unable to update runner: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Error updating runner",
			fmt.Sprintf("API returned status %d when updating runner", httpResp.StatusCode),
		)
		return
	}

	var apiResponse struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Description    string `json:"description"`
		ResourceClass  string `json:"resource_class"`
		Platform       string `json:"platform"`
		IP             string `json:"ip"`
		Hostname       string `json:"hostname"`
		Version        string `json:"version"`
		FirstConnected string `json:"first_connected"`
		LastConnected  string `json:"last_connected"`
		LastUsed       string `json:"last_used"`
		State          string `json:"state"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error parsing response",
			fmt.Sprintf("Unable to parse runner update response: %v", err),
		)
		return
	}

	// Update the model with response data
	data.ID = types.StringValue(apiResponse.ID)
	data.Name = types.StringValue(apiResponse.Name)
	data.Description = types.StringValue(apiResponse.Description)
	data.ResourceClass = types.StringValue(apiResponse.ResourceClass)
	data.Platform = types.StringValue(apiResponse.Platform)
	data.IP = types.StringValue(apiResponse.IP)
	data.Hostname = types.StringValue(apiResponse.Hostname)
	data.Version = types.StringValue(apiResponse.Version)
	data.FirstConnected = types.StringValue(apiResponse.FirstConnected)
	data.LastConnected = types.StringValue(apiResponse.LastConnected)
	data.LastUsed = types.StringValue(apiResponse.LastUsed)
	data.State = types.StringValue(apiResponse.State)

	tflog.Trace(ctx, "updated runner resource", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RunnerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RunnerResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Make API request to delete runner
	url := fmt.Sprintf("%s/runner/%s", r.client.BaseURL, data.ID.ValueString())
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating request",
			fmt.Sprintf("Unable to create runner deletion request: %v", err),
		)
		return
	}

	httpReq.Header.Set("Circle-Token", r.client.ApiToken)

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting runner",
			fmt.Sprintf("Unable to delete runner: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting runner",
			fmt.Sprintf("API returned status %d when deleting runner", httpResp.StatusCode),
		)
		return
	}

	tflog.Trace(ctx, "deleted runner resource", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
}

func (r *RunnerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
