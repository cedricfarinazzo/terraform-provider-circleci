package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &PoliciesDataSource{}

func NewPoliciesDataSource() datasource.DataSource {
	return &PoliciesDataSource{}
}

type PoliciesDataSource struct {
	client *CircleCIClient
}

type PoliciesDataSourceModel struct {
	OrgID    types.String      `tfsdk:"org_id"`
	Policies []PolicyDataModel `tfsdk:"policies"`
}

type PolicyDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Content     types.String `tfsdk:"content"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// CircleCI API models for policies list
type PoliciesListAPI struct {
	Items         []PolicyAPI `json:"items"`
	NextPageToken string      `json:"next_page_token"`
}

func (d *PoliciesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policies"
}

func (d *PoliciesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CircleCI Policies data source. Retrieves a list of all policies in an organization.",

		Attributes: map[string]schema.Attribute{
			"org_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The organization ID for which to retrieve policies.",
			},
			"policies": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of policies in the organization.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the policy.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the policy.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The description of the policy.",
						},
						"content": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The policy content in OPA Rego format.",
						},
						"enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the policy is enabled.",
						},
						"created_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The date and time the policy was created.",
						},
						"updated_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The date and time the policy was last updated.",
						},
					},
				},
			},
		},
	}
}

func (d *PoliciesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PoliciesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrgID.ValueString()
	endpoint := fmt.Sprintf("/v2/policy/%s", orgID)

	httpResp, err := d.client.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get policies", err.Error())
		return
	}
	defer httpResp.Body.Close()

	var policiesResponse PoliciesListAPI
	if err := json.NewDecoder(httpResp.Body).Decode(&policiesResponse); err != nil {
		resp.Diagnostics.AddError("Failed to decode response", err.Error())
		return
	}

	// Convert API response to data source model
	policies := make([]PolicyDataModel, len(policiesResponse.Items))
	for i, policy := range policiesResponse.Items {
		policies[i] = PolicyDataModel{
			ID:          types.StringValue(policy.ID),
			Name:        types.StringValue(policy.Name),
			Description: types.StringValue(policy.Description),
			Content:     types.StringValue(policy.Content),
			Enabled:     types.BoolValue(policy.Enabled),
			CreatedAt:   types.StringValue(policy.CreatedAt),
			UpdatedAt:   types.StringValue(policy.UpdatedAt),
		}
	}

	data.Policies = policies

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
