package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the provider satisfies the framework interface.
var _ provider.Provider = &jooplerProvider{}

type jooplerProvider struct {
	version string
}

type providerModel struct {
	APIURL types.String `tfsdk:"api_url"`
	APIKey types.String `tfsdk:"api_key"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &jooplerProvider{version: version}
	}
}

func (p *jooplerProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "joopler"
	resp.Version = p.version
}

func (p *jooplerProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage the scaffolding of a Joopler compliance program as code. " +
			"Sign-off and evidence attestation are human acts and are intentionally not managed by this provider.",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Joopler API base URL. Defaults to `https://api.joopler.com` or `JOOPLER_API_URL`.",
			},
			"api_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Write-scoped tenant API key (`jpl_...`). Defaults to `JOOPLER_API_KEY`. " +
					"Mint one at https://app.joopler.com/developers - it requires an existing Joopler workspace.",
			},
		},
	}
}

func (p *jooplerProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiURL := firstNonEmpty(cfg.APIURL.ValueString(), os.Getenv("JOOPLER_API_URL"), "https://api.joopler.com")
	apiKey := firstNonEmpty(cfg.APIKey.ValueString(), os.Getenv("JOOPLER_API_KEY"))
	if apiKey == "" {
		// Say where a key comes from, and that it comes from an ACCOUNT. Anyone
		// reaching this has already had `terraform init` succeed - the provider is
		// public on the registry - so the failure reads like a misconfiguration when
		// the truth is they need a workspace first. Naming the page without saying
		// that sends them to a login they may not be able to create.
		resp.Diagnostics.AddError(
			"Missing Joopler API key",
			"This provider manages an existing Joopler workspace, so it needs a write-scoped API key (jpl_...).\n\n"+
				"Set it in the provider block as `api_key`, or in the JOOPLER_API_KEY environment variable.\n\n"+
				"If you already have a workspace, mint a key at https://app.joopler.com/developers.\n"+
				"If you do not have one yet, Joopler is currently invitation-only - talk to us at "+
				"https://joopler.com and we will get you set up.",
		)
		return
	}

	client := NewClient(apiURL, apiKey)
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *jooplerProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewControlOwnerResource,
		NewPolicyOwnerResource,
		NewVendorResource,
		NewAuditTargetResource,
		NewConnectorResource,
	}
}

func (p *jooplerProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewControlStatusDataSource,
		NewEnvironmentDataSource,
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
