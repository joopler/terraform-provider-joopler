package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &controlStatusDataSource{}

type controlStatusDataSource struct{ client *Client }

type controlStatusModel struct {
	Statuses types.Map   `tfsdk:"statuses"`
	Passing  types.Int64 `tfsdk:"passing"`
	Failing  types.Int64 `tfsdk:"failing"`
	Total    types.Int64 `tfsdk:"total"`
}

func NewControlStatusDataSource() datasource.DataSource { return &controlStatusDataSource{} }

func (d *controlStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_control_status"
}

func (d *controlStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Live posture: each control's current status plus rollup counts.",
		Attributes: map[string]schema.Attribute{
			"statuses": schema.MapAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Control key -> status (PASS/FAIL/ERROR/PENDING)."},
			"passing":  schema.Int64Attribute{Computed: true, MarkdownDescription: "Count of passing controls."},
			"failing":  schema.Int64Attribute{Computed: true, MarkdownDescription: "Count of failing controls."},
			"total":    schema.Int64Attribute{Computed: true, MarkdownDescription: "Total controls."},
		},
	}
}

func (d *controlStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*Client)
	}
}

func (d *controlStatusDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var out struct {
		Controls []struct {
			Key    string `json:"key"`
			Status string `json:"status"`
		} `json:"controls"`
	}
	if err := d.client.Get(ctx, "/v1/control-status", &out); err != nil {
		resp.Diagnostics.AddError("Read control status failed", err.Error())
		return
	}

	statuses := make(map[string]string, len(out.Controls))
	var passing, failing int64
	for _, c := range out.Controls {
		statuses[c.Key] = c.Status
		switch c.Status {
		case "PASS":
			passing++
		case "FAIL":
			failing++
		}
	}

	var state controlStatusModel
	state.Statuses, _ = types.MapValueFrom(ctx, types.StringType, statuses)
	state.Passing = types.Int64Value(passing)
	state.Failing = types.Int64Value(failing)
	state.Total = types.Int64Value(int64(len(out.Controls)))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
