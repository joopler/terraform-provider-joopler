package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &connectorResource{}

type connectorResource struct{ client *Client }

type connectorModel struct {
	Key     types.String `tfsdk:"key"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Config  types.Map    `tfsdk:"config"`
}

func NewConnectorResource() resource.Resource { return &connectorResource{} }

func (r *connectorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector"
}

func (r *connectorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A connector's non-secret configuration. Secrets are set separately and never pass through Terraform state.",
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Connector key (e.g. `aws`, `okta`, `github`).",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the connector is enabled. Defaults to true.",
			},
			"config": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Non-secret config (e.g. `region`, `orgUrl`). Never put secrets here.",
			},
		},
	}
}

func (r *connectorResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*Client)
	}
}

func (r *connectorResource) apply(ctx context.Context, m connectorModel) error {
	config := map[string]string{}
	if !m.Config.IsNull() && !m.Config.IsUnknown() {
		m.Config.ElementsAs(ctx, &config, false)
	}
	enabled := true
	if !m.Enabled.IsNull() && !m.Enabled.IsUnknown() {
		enabled = m.Enabled.ValueBool()
	}
	return r.client.Put(ctx, fmt.Sprintf("/v1/integrations/%s", url.PathEscape(m.Key.ValueString())),
		map[string]any{"enabled": enabled, "config": config}, nil)
}

func (r *connectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan connectorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Configure connector failed", err.Error())
		return
	}
	if plan.Enabled.IsNull() || plan.Enabled.IsUnknown() {
		plan.Enabled = types.BoolValue(true)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *connectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state connectorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out struct {
		Connectors []struct {
			Key     string `json:"key"`
			Enabled bool   `json:"enabled"`
		} `json:"connectors"`
	}
	if err := r.client.Get(ctx, "/v1/integrations", &out); err != nil {
		// A missing list endpoint should not corrupt state; keep the plan values.
		return
	}
	for _, c := range out.Connectors {
		if c.Key == state.Key.ValueString() {
			state.Enabled = types.BoolValue(c.Enabled)
			resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *connectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan connectorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Update connector failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *connectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state connectorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Disable the connector on delete (config is left in place).
	state.Enabled = types.BoolValue(false)
	if err := r.apply(ctx, state); err != nil {
		resp.Diagnostics.AddError("Disable connector failed", err.Error())
	}
}
