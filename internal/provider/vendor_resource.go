package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &vendorResource{}

type vendorResource struct{ client *Client }

type vendorModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Category   types.String `tfsdk:"category"`
	RiskTier   types.String `tfsdk:"risk_tier"`
	DataAccess types.String `tfsdk:"data_access"`
	OwnerEmail types.String `tfsdk:"owner_email"`
	Status     types.String `tfsdk:"status"`
	Website    types.String `tfsdk:"website"`
	Notes      types.String `tfsdk:"notes"`
}

func NewVendorResource() resource.Resource { return &vendorResource{} }

func (r *vendorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vendor"
}

func (r *vendorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A vendor / subprocessor in the TPRM register. Deleting the resource offboards the vendor (Joopler keeps the record for the audit trail).",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Vendor id."},
			"name":        schema.StringAttribute{Required: true, MarkdownDescription: "Vendor name."},
			"category":    schema.StringAttribute{Optional: true, MarkdownDescription: "Category (e.g. Payments)."},
			"risk_tier":   schema.StringAttribute{Optional: true, MarkdownDescription: "critical | high | medium | low."},
			"data_access": schema.StringAttribute{Optional: true, MarkdownDescription: "none | internal | confidential | pii | phi."},
			"owner_email": schema.StringAttribute{Optional: true, MarkdownDescription: "Vendor owner email."},
			"status":      schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "in_review | approved | offboarded."},
			"website":     schema.StringAttribute{Optional: true, MarkdownDescription: "Vendor website."},
			"notes":       schema.StringAttribute{Optional: true, MarkdownDescription: "Free-text notes."},
		},
	}
}

func (r *vendorResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*Client)
	}
}

// body builds the mutable field map, omitting unset optionals.
func (m vendorModel) body() map[string]any {
	b := map[string]any{"name": m.Name.ValueString()}
	put := func(k string, v types.String) {
		if !v.IsNull() && !v.IsUnknown() {
			b[k] = v.ValueString()
		}
	}
	put("category", m.Category)
	put("riskTier", m.RiskTier)
	put("dataAccess", m.DataAccess)
	put("ownerEmail", m.OwnerEmail)
	put("status", m.Status)
	put("website", m.Website)
	put("notes", m.Notes)
	return b
}

func (r *vendorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vendorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out struct {
		ID     string `json:"id"`
		Vendor struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"vendor"`
	}
	if err := r.client.Post(ctx, "/v1/vendors", plan.body(), &out); err != nil {
		resp.Diagnostics.AddError("Create vendor failed", err.Error())
		return
	}
	id := out.ID
	if id == "" {
		id = out.Vendor.ID
	}
	plan.ID = types.StringValue(id)
	if plan.Status.IsNull() || plan.Status.IsUnknown() {
		plan.Status = types.StringValue(firstNonEmpty(out.Vendor.Status, "in_review"))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *vendorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vendorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out struct {
		Vendors []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			Category   string `json:"category"`
			RiskTier   string `json:"riskTier"`
			DataAccess string `json:"dataAccess"`
			OwnerEmail string `json:"ownerEmail"`
			Status     string `json:"status"`
			Website    string `json:"website"`
			Notes      string `json:"notes"`
		} `json:"vendors"`
	}
	if err := r.client.Get(ctx, "/v1/vendors", &out); err != nil {
		resp.Diagnostics.AddError("Read vendor failed", err.Error())
		return
	}
	for _, v := range out.Vendors {
		if v.ID == state.ID.ValueString() {
			state.Name = types.StringValue(v.Name)
			state.Status = types.StringValue(v.Status)
			setOpt(&state.Category, v.Category)
			setOpt(&state.RiskTier, v.RiskTier)
			setOpt(&state.DataAccess, v.DataAccess)
			setOpt(&state.OwnerEmail, v.OwnerEmail)
			setOpt(&state.Website, v.Website)
			setOpt(&state.Notes, v.Notes)
			resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *vendorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vendorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state vendorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	if err := r.client.Put(ctx, fmt.Sprintf("/v1/vendors/%s", url.PathEscape(state.ID.ValueString())), plan.body(), nil); err != nil {
		resp.Diagnostics.AddError("Update vendor failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *vendorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vendorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// No hard delete: offboard so the subprocessor record is retained for audit.
	if err := r.client.Put(ctx, fmt.Sprintf("/v1/vendors/%s", url.PathEscape(state.ID.ValueString())),
		map[string]any{"status": "offboarded"}, nil); err != nil {
		resp.Diagnostics.AddError("Offboard vendor failed", err.Error())
	}
}

// setOpt keeps an optional attribute null when the API returns empty, so the
// plan stays stable for fields the user never set.
func setOpt(dst *types.String, v string) {
	if v == "" {
		if dst.IsUnknown() {
			*dst = types.StringNull()
		}
		return
	}
	*dst = types.StringValue(v)
}
