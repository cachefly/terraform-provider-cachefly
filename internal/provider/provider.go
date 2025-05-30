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

	"github.com/avvvet/terraform-provider-cachefly/internal/provider/datasources"
	"github.com/avvvet/terraform-provider-cachefly/internal/provider/resources"
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
		MarkdownDescription: `
# CacheFly Terraform Provider

The CacheFly provider allows you to manage CacheFly CDN resources using Terraform.

Use the CacheFly provider to manage services, accounts, users, and other CDN configurations.

## Authentication

The provider requires an API token to authenticate with the CacheFly API. 

### API Token

You can provide the API token in several ways:

1. **Provider configuration** (recommended for development):
` + "```hcl" + `
provider "cachefly" {
  api_token = "your-api-token-here"
}
` + "```" + `

2. **Environment variable** (recommended for production):
` + "```bash" + `
export CACHEFLY_API_TOKEN="your-api-token-here"
` + "```" + `

3. **Terraform variables**:
` + "```hcl" + `
variable "cachefly_api_token" {
  description = "CacheFly API Token"
  type        = string
  sensitive   = true
}

provider "cachefly" {
  api_token = var.cachefly_api_token
}
` + "```" + `

## Example Usage

` + "```hcl" + `
terraform {
  required_providers {
    cachefly = {
      source  = "cachefly/cachefly"
      version = "~> 1.0"
    }
  }
}

# Configure the CacheFly Provider
provider "cachefly" {
  api_token = var.cachefly_api_token
}

# Create a service
resource "cachefly_service" "example" {
  name        = "my-cdn-service"
  unique_name = "my-unique-service"
  description = "Example CDN service"
  auto_ssl    = true
}
` + "```" + `
		`,
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

	// Log configuration (without sensitive data)
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

	// Create our provider client wrapper
	client := cacheflyClient

	// Make the client available to resources and data sources
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Successfully configured CacheFly provider", map[string]interface{}{
		"base_url": baseURL,
	})
}

func (p *CacheFlyProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Service resources
		resources.NewServiceResource,

		// Add other resources here as you implement them
		// resources.NewServiceDomainResource,
		// resources.NewServiceRuleResource,
		// resources.NewServiceOptionsResource,
		// resources.NewCertificateResource,
		// resources.NewOriginResource,
		// resources.NewUserResource,
		// resources.NewAccountResource,
		// resources.NewScriptConfigResource,
		// resources.NewTLSProfileResource,
	}
}

func (p *CacheFlyProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Service data sources
		datasources.NewServiceDataSource,
		// datasources.NewServicesDataSource, // For listing multiple services - not implemented yet

		// Add other data sources here as you implement them
		// datasources.NewServiceDomainsDataSource,
		// datasources.NewServiceRulesDataSource,
		// datasources.NewCertificatesDataSource,
		// datasources.NewOriginsDataSource,
		// datasources.NewUsersDataSource,
		// datasources.NewAccountsDataSource,
		// datasources.NewTLSProfilesDataSource,
	}
}

// New returns a function that creates a new provider instance
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CacheFlyProvider{
			version: version,
		}
	}
}

// Helper function to get configuration values with fallback to environment variables
func getConfigValue(configValue types.String, envVar, defaultValue string) string {
	// If explicitly set in config, use that
	if !configValue.IsNull() && !configValue.IsUnknown() {
		return configValue.ValueString()
	}

	// Otherwise, try environment variable
	if envValue := os.Getenv(envVar); envValue != "" {
		return envValue
	}

	// Finally, use default value
	return defaultValue
}

// Placeholder functions for resources and data sources
// These will be replaced when you implement the actual resources

func NewServiceResource() resource.Resource {
	// This will be implemented in resources/service_resource.go
	panic("NewServiceResource not yet implemented")
}

func NewServiceDataSource() datasource.DataSource {
	// This will be implemented in datasources/service_data_source.go
	panic("NewServiceDataSource not yet implemented")
}

func NewServicesDataSource() datasource.DataSource {
	// This will be implemented in datasources/services_data_source.go
	panic("NewServicesDataSource not yet implemented")
}
