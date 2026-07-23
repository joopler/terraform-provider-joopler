package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &policyOwnerResource{}

type policyOwnerResource struct{ client *Client }

type policyOwnerModel struct {
	PolicyKey types.String `tfsdk:"policy_key"`
	Owner     types.String `tfsdk:"owner"`
}

func NewPolicyOwnerResource() resource.Resource { return &policyOwnerResource{} }

func (r *policyOwnerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_owner"
}

func (r *policyOwnerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Assign the owner of a policy. The owner (or an admin) still signs off on the policy as a human act - this only assigns accountability.",
		Attributes: map[string]schema.Attribute{
			"policy_key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Policy key (e.g. `acceptable-use`).",
			},
			"owner": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Email of the accountable owner.",
			},
		},
	}
}

func (r *policyOwnerResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*Client)
	}
}

func (r *policyOwnerResource) set(ctx context.Context, key string, owner any) error {
	return r.client.Put(ctx, fmt.Sprintf("/v1/policies/%s/owner", url.PathEscape(key)),
		map[string]any{"owner": owner}, nil)
}

func (r *policyOwnerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan policyOwnerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.set(ctx, plan.PolicyKey.ValueString(), plan.Owner.ValueString()); err != nil {
		resp.Diagnostics.AddError("Assign policy owner failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *policyOwnerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state policyOwnerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out struct {
		Policies []struct {
			Key   string `json:"key"`
			Owner string `json:"owner"`
		} `json:"policies"`
	}
	if err := r.client.Get(ctx, "/v1/policies", &out); err != nil {
		resp.Diagnostics.AddError("Read policy owner failed", err.Error())
		return
	}
	for _, p := range out.Policies {
		if p.Key == state.PolicyKey.ValueString() {
			if p.Owner == "" {
				resp.State.RemoveResource(ctx)
				return
			}
			state.Owner = types.StringValue(p.Owner)
			resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *policyOwnerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan policyOwnerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.set(ctx, plan.PolicyKey.ValueString(), plan.Owner.ValueString()); err != nil {
		resp.Diagnostics.AddError("Update policy owner failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *policyOwnerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state policyOwnerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.set(ctx, state.PolicyKey.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("Clear policy owner failed", err.Error())
	}
}
