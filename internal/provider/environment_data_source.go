package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &environmentDataSource{}

type environmentDataSource struct{ client *Client }

type environmentModel struct {
	ConnectedSystems  types.List  `tfsdk:"connected_systems"`
	IdentityProviders types.List  `tfsdk:"identity_providers"`
	EndpointMgmt      types.List  `tfsdk:"endpoint_management"`
	SubprocessorCount types.Int64 `tfsdk:"subprocessor_count"`
	WorkforceCount    types.Int64 `tfsdk:"workforce_count"`
}

func NewEnvironmentDataSource() datasource.DataSource { return &environmentDataSource{} }

func (d *environmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *environmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The tenant's connected environment - the observed reality Joopler grounds policies and the assistant in.",
		Attributes: map[string]schema.Attribute{
			"connected_systems":  schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Labels of connected systems."},
			"identity_providers": schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Connected identity providers."},
			"endpoint_management": schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Connected device-management tools."},
			"subprocessor_count": schema.Int64Attribute{Computed: true, MarkdownDescription: "Number of subprocessors on file."},
			"workforce_count":    schema.Int64Attribute{Computed: true, MarkdownDescription: "People on the roster."},
		},
	}
}

func (d *environmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*Client)
	}
}

func (d *environmentDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var out struct {
		ConnectedSystems []struct {
			Label string `json:"label"`
		} `json:"connectedSystems"`
		IdentityProviders  []string `json:"identityProviders"`
		EndpointManagement []string `json:"endpointManagement"`
		Subprocessors      []struct {
			Name string `json:"name"`
		} `json:"subprocessors"`
		WorkforceCount int64 `json:"workforceCount"`
	}
	if err := d.client.Get(ctx, "/v1/environment", &out); err != nil {
		resp.Diagnostics.AddError("Read environment failed", err.Error())
		return
	}

	labels := make([]string, 0, len(out.ConnectedSystems))
	for _, s := range out.ConnectedSystems {
		labels = append(labels, s.Label)
	}

	var state environmentModel
	state.ConnectedSystems, _ = types.ListValueFrom(ctx, types.StringType, labels)
	state.IdentityProviders, _ = types.ListValueFrom(ctx, types.StringType, out.IdentityProviders)
	state.EndpointMgmt, _ = types.ListValueFrom(ctx, types.StringType, out.EndpointManagement)
	state.SubprocessorCount = types.Int64Value(int64(len(out.Subprocessors)))
	state.WorkforceCount = types.Int64Value(out.WorkforceCount)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
