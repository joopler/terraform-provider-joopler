package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &controlOwnerResource{}

type controlOwnerResource struct{ client *Client }

type controlOwnerModel struct {
	ControlKey types.String `tfsdk:"control_key"`
	OwnerEmail types.String `tfsdk:"owner_email"`
}

func NewControlOwnerResource() resource.Resource { return &controlOwnerResource{} }

func (r *controlOwnerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_control_owner"
}

func (r *controlOwnerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Assign the accountable owner of a control.",
		Attributes: map[string]schema.Attribute{
			"control_key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Control key (e.g. `aws-cloudtrail-enabled`).",
			},
			"owner_email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Email of the accountable owner.",
			},
		},
	}
}

func (r *controlOwnerResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*Client)
	}
}

func (r *controlOwnerResource) set(ctx context.Context, key, email string) error {
	return r.client.Put(ctx, fmt.Sprintf("/v1/controls/%s/owner", url.PathEscape(key)),
		map[string]any{"ownerEmail": email}, nil)
}

func (r *controlOwnerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan controlOwnerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.set(ctx, plan.ControlKey.ValueString(), plan.OwnerEmail.ValueString()); err != nil {
		resp.Diagnostics.AddError("Assign control owner failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *controlOwnerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state controlOwnerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out struct {
		Controls []struct {
			Key        string `json:"key"`
			OwnerEmail string `json:"ownerEmail"`
		} `json:"controls"`
	}
	if err := r.client.Get(ctx, "/v1/control-status", &out); err != nil {
		resp.Diagnostics.AddError("Read control owner failed", err.Error())
		return
	}
	for _, c := range out.Controls {
		if c.Key == state.ControlKey.ValueString() {
			if c.OwnerEmail == "" {
				resp.State.RemoveResource(ctx) // owner cleared out of band
				return
			}
			state.OwnerEmail = types.StringValue(c.OwnerEmail)
			resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
			return
		}
	}
	resp.State.RemoveResource(ctx) // control no longer exists
}

func (r *controlOwnerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan controlOwnerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.set(ctx, plan.ControlKey.ValueString(), plan.OwnerEmail.ValueString()); err != nil {
		resp.Diagnostics.AddError("Update control owner failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *controlOwnerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state controlOwnerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Clearing the owner is the delete: PUT a null owner.
	if err := r.client.Put(ctx, fmt.Sprintf("/v1/controls/%s/owner", url.PathEscape(state.ControlKey.ValueString())),
		map[string]any{"ownerEmail": nil}, nil); err != nil {
		resp.Diagnostics.AddError("Clear control owner failed", err.Error())
	}
}
