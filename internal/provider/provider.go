package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/cachefly/cachefly-go-sdk/pkg/cachefly"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/datasources"
	"github.com/cachefly/terraform-provider-cachefly/internal/provider/resources"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ provider.Provider = &CacheFlyProvider{}

// CacheFlyProvider defines the provider implementation.
type CacheFlyProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// CacheFlyProviderModel describes the provider data model.
type CacheFlyProviderModel struct {
	APIToken types.String `tfsdk:"api_token"`
	BaseURL  types.String `tfsdk:"base_url"`
}

// CacheFlyClient holds the SDK client with all service APIs
type CacheFlyClient struct {
	// Main SDK client with all services
	Client *cachefly.Client

	// Configuration for easy access
	APIToken string
	BaseURL  string
}

func (p *CacheFlyProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cachefly"
	resp.Version = p.version
}

func (p *CacheFlyProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The CacheFly provider allows you to manage CacheFly CDN resources using Terraform.",
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				MarkdownDescription: "The API token for authenticating with CacheFly. Can also be set with the `CACHEFLY_API_TOKEN` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "The base URL for the CacheFly API. Defaults to `https://api.cachefly.com/api/2.5`. Can also be set with the `CACHEFLY_BASE_URL` environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *CacheFlyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config CacheFlyProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default values and get from environment if not set
	apiToken := getConfigValue(config.APIToken, "CACHEFLY_API_TOKEN", "")
	baseURL := getConfigValue(config.BaseURL, "CACHEFLY_BASE_URL", "https://api.cachefly.com/api/2.5")

	// Validate required configuration
	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing API Token",
			"The CacheFly API token is required but was not found. "+
				"Please set it in the provider configuration or via the CACHEFLY_API_TOKEN environment variable.",
		)
		return
	}

	// Log configuration
	tflog.Debug(ctx, "Configuring CacheFly provider", map[string]interface{}{
		"base_url": baseURL,
		"version":  p.version,
	})

	// Create CacheFly SDK client using the proper constructor
	cacheflyClient := cachefly.NewClient(
		cachefly.WithToken(apiToken),
		cachefly.WithBaseURL(baseURL),
	)

	if cacheflyClient == nil {
		resp.Diagnostics.AddError(
			"Failed to Create CacheFly Client",
			"Unable to create CacheFly SDK client. Please check your configuration.",
		)
		return
	}

	client := cacheflyClient

	// client available to resources and data sources
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Successfully configured CacheFly provider", map[string]interface{}{
		"base_url": baseURL,
	})
}

func (p *CacheFlyProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewServiceResource,
		resources.NewServiceDomainResource,
		resources.NewOriginResource,
		resources.NewServiceOptionsResource,
		resources.NewUserResource,
		resources.NewScriptConfigResource,
	}
}

func (p *CacheFlyProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewServiceDataSource,
		datasources.NewServiceDomainDataSource,
		datasources.NewServiceDomainsDataSource,
		datasources.NewOriginDataSource,
		datasources.NewOriginsDataSource,
		datasources.NewServiceOptionsDataSource,
		datasources.NewUsersDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CacheFlyProvider{
			version: version,
		}
	}
}

func getConfigValue(configValue types.String, envVar, defaultValue string) string {

	if !configValue.IsNull() && !configValue.IsUnknown() {
		return configValue.ValueString()
	}

	if envValue := os.Getenv(envVar); envValue != "" {
		return envValue
	}

	return defaultValue
}
