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

// Ensure NodeRushProvider satisfies the provider.Provider interface.
var _ provider.Provider = &NodeRushProvider{}

type NodeRushProvider struct {
	version string
}

type providerModel struct {
	APIURL   types.String `tfsdk:"api_url"`
	APIToken types.String `tfsdk:"api_token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &NodeRushProvider{version: version}
	}
}

func (p *NodeRushProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "noderush"
	resp.Version = p.version
}

func (p *NodeRushProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage NodeRush resources (VPS nodes, block volumes, SSH keys) through the NodeRush API.",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				MarkdownDescription: "Base URL of the NodeRush API. Defaults to the `NODERUSH_API_URL` env var, or `https://api.noderush.io`.",
				Optional:            true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "A NodeRush personal access token. Defaults to the `NODERUSH_API_TOKEN` env var. Required.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *NodeRushProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiURL := os.Getenv("NODERUSH_API_URL")
	if !cfg.APIURL.IsNull() {
		apiURL = cfg.APIURL.ValueString()
	}
	if apiURL == "" {
		apiURL = "https://api.noderush.io"
	}

	token := os.Getenv("NODERUSH_API_TOKEN")
	if !cfg.APIToken.IsNull() {
		token = cfg.APIToken.ValueString()
	}
	if token == "" {
		resp.Diagnostics.AddError(
			"Missing NodeRush API token",
			"Set the provider `api_token` attribute or the NODERUSH_API_TOKEN environment variable.",
		)
		return
	}

	client := NewClient(apiURL, token)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *NodeRushProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNodeResource,
		NewSSHKeyResource,
		NewVolumeResource,
	}
}

func (p *NodeRushProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRegionsDataSource,
		NewImagesDataSource,
		NewPlansDataSource,
	}
}
