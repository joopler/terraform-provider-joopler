package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &auditTargetResource{}

type auditTargetResource struct{ client *Client }

type auditTargetModel struct {
	TargetDate types.String `tfsdk:"target_date"`
}

func NewAuditTargetResource() resource.Resource { return &auditTargetResource{} }

func (r *auditTargetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_target"
}

func (r *auditTargetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The tenant's audit-readiness target date (singleton). Drives the countdown and journey nudges.",
		Attributes: map[string]schema.Attribute{
			"target_date": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Target date, ISO `YYYY-MM-DD`.",
			},
		},
	}
}

func (r *auditTargetResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*Client)
	}
}

func (r *auditTargetResource) set(ctx context.Context, date any) error {
	return r.client.Put(ctx, "/v1/audit-target", map[string]any{"targetDate": date}, nil)
}

func (r *auditTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan auditTargetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.set(ctx, plan.TargetDate.ValueString()); err != nil {
		resp.Diagnostics.AddError("Set audit target failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *auditTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state auditTargetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out struct {
		TargetDate *string `json:"targetDate"`
	}
	if err := r.client.Get(ctx, "/v1/audit-target", &out); err != nil {
		resp.Diagnostics.AddError("Read audit target failed", err.Error())
		return
	}
	if out.TargetDate == nil || *out.TargetDate == "" {
		resp.State.RemoveResource(ctx)
		return
	}
	// Normalize to the date portion so a returned timestamp does not thrash the plan.
	d := *out.TargetDate
	if len(d) >= 10 {
		d = d[:10]
	}
	state.TargetDate = types.StringValue(d)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *auditTargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan auditTargetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.set(ctx, plan.TargetDate.ValueString()); err != nil {
		resp.Diagnostics.AddError("Update audit target failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *auditTargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.set(ctx, nil); err != nil {
		resp.Diagnostics.AddError("Clear audit target failed", err.Error())
	}
}
