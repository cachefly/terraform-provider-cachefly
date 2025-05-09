package provider

import (
	"context"

	"github.com/avvvet/cachefly-sdk-go/pkg/cachefly"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &cacheflyProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &cacheflyProvider{
			version: version,
		}
	}
}

// cacheflyProvider is the provider implementation.
type cacheflyProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
	client  *cachefly.Client // Hold the cachefly client
}

// Metadata returns the provider type name.
func (p *cacheflyProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cachefly"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *cacheflyProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Required:    true,
				Description: "The API token for authenticating with the CacheFly API yellow.",
			},
		},
	}
}

// Configure prepares a cachefly API client for data sources and resources.
func (p *cacheflyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve the api_token from the configuration
	var config struct {
		ApiToken types.String `tfsdk:"api_token"`
	}

	// Read provider configuration (api_token)
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ApiToken.IsNull() || config.ApiToken.ValueString() == "" {
		resp.Diagnostics.AddError("Configuration Error", "The api_token must be provided.")
		return
	}

	client := cachefly.NewClient(
		cachefly.WithToken(config.ApiToken.ValueString()),
	)

	p.client = client

}

// DataSources defines the data sources implemented in the provider.
func (p *cacheflyProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource { return &servicesDataSource{client: p.client} },
	}
}

// Resources defines the resources implemented in the provider.
func (p *cacheflyProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
