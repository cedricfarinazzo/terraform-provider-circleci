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
)

var _ resource.Resource = &UsageExportResource{}
var _ resource.ResourceWithImportState = &UsageExportResource{}

func NewUsageExportResource() resource.Resource {
	return &UsageExportResource{}
}

type UsageExportResource struct {
	client *CircleCIClient
}

type UsageExportResourceModel struct {
	ID          types.String `tfsdk:"id"`
	OrgID       types.String `tfsdk:"org_id"`
	Start       types.String `tfsdk:"start"`
	End         types.String `tfsdk:"end"`
	Status      types.String `tfsdk:"status"`
	DownloadURL types.String `tfsdk:"download_url"`
	CreatedAt   types.String `tfsdk:"created_at"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
}

// CircleCI API models for usage exports
type UsageExportAPI struct {
	ID          string `json:"id"`
	OrgID       string `json:"org_id"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Status      string `json:"status"`
	DownloadURL string `json:"download_url,omitempty"`
	CreatedAt   string `json:"created_at"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

type UsageExportRequest struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

func (r *UsageExportResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_usage_export"
}

func (r *UsageExportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Usage Export resource. This resource allows you to export organization usage data for a specified time range.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the usage export.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The organization ID for which to export usage data.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"start": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The start date for the usage export (ISO 8601 format).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"end": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The end date for the usage export (ISO 8601 format).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the usage export (pending, processing, completed, failed).",
			},
			"download_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The download URL for the completed export (available when status is 'completed').",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time the export was created.",
			},
			"expires_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time the export download will expire.",
			},
		},
	}
}

func (r *UsageExportResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UsageExportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UsageExportResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgID.ValueString()
	exportRequest := UsageExportRequest{
		Start: data.Start.ValueString(),
		End:   data.End.ValueString(),
	}

	endpoint := fmt.Sprintf("/v2/organization/%s/usage-export", orgID)
	httpResp, err := r.client.MakeRequest(ctx, "POST", endpoint, exportRequest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create usage export", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Failed to create usage export",
			fmt.Sprintf("CircleCI API returned status %d", httpResp.StatusCode),
		)
		return
	}

	var export UsageExportAPI
	if err := json.NewDecoder(httpResp.Body).Decode(&export); err != nil {
		resp.Diagnostics.AddError("Failed to decode response", err.Error())
		return
	}

	r.mapExportToModel(&export, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UsageExportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UsageExportResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgID.ValueString()
	exportID := data.ID.ValueString()

	endpoint := fmt.Sprintf("/v2/organization/%s/usage-export/%s", orgID, exportID)
	httpResp, err := r.client.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get usage export", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Failed to get usage export",
			fmt.Sprintf("CircleCI API returned status %d", httpResp.StatusCode),
		)
		return
	}

	var export UsageExportAPI
	if err := json.NewDecoder(httpResp.Body).Decode(&export); err != nil {
		resp.Diagnostics.AddError("Failed to decode response", err.Error())
		return
	}

	r.mapExportToModel(&export, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UsageExportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Usage exports are immutable once created
	// Just refresh the current state
	r.Read(ctx, resource.ReadRequest{State: req.State}, &resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics})
}

func (r *UsageExportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UsageExportResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgID.ValueString()
	exportID := data.ID.ValueString()

	endpoint := fmt.Sprintf("/v2/organization/%s/usage-export/%s", orgID, exportID)
	httpResp, err := r.client.MakeRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete usage export", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Failed to delete usage export",
			fmt.Sprintf("CircleCI API returned status %d", httpResp.StatusCode),
		)
		return
	}
}

func (r *UsageExportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "org_id:export_id"
	// Example: "bb604b45-b6b0-4b81-ad80-796f15eddf87:550e8400-e29b-41d4-a716-446655440000"
	idParts := strings.SplitN(req.ID, ":", 2)
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Import ID must be in format 'org_id:export_id', got: %q", req.ID),
		)
		return
	}

	orgID := idParts[0]
	exportID := idParts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), exportID)...)
}

func (r *UsageExportResource) mapExportToModel(export *UsageExportAPI, data *UsageExportResourceModel) {
	data.ID = types.StringValue(export.ID)
	data.OrgID = types.StringValue(export.OrgID)
	data.Start = types.StringValue(export.Start)
	data.End = types.StringValue(export.End)
	data.Status = types.StringValue(export.Status)
	data.CreatedAt = types.StringValue(export.CreatedAt)

	if export.DownloadURL != "" {
		data.DownloadURL = types.StringValue(export.DownloadURL)
	} else {
		data.DownloadURL = types.StringNull()
	}

	if export.ExpiresAt != "" {
		data.ExpiresAt = types.StringValue(export.ExpiresAt)
	} else {
		data.ExpiresAt = types.StringNull()
	}
}
