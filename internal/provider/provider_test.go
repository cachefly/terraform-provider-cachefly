package provider

import (
	"context"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cachefly": providerserver.NewProtocol6WithError(New("test")()),
}

func TestProvider(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig,
			},
		},
	})
}

const testAccProviderConfig = `
provider "cachefly" {
  api_token = "test-token"
  base_url  = "https://api.test.cachefly.com/api/2.5"
}
`

// TestProviderSchema tests the provider schema validation
func TestProviderSchema(t *testing.T) {
	ctx := context.Background()
	p := &CacheFlyProvider{version: "test"}

	// Get the provider schema
	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}

	p.Schema(ctx, schemaReq, schemaResp)

	// Verify no errors in schema
	assert.False(t, schemaResp.Diagnostics.HasError(), "Provider schema should not have errors")

	// Verify expected attributes exist
	attrs := schemaResp.Schema.Attributes
	assert.Contains(t, attrs, "api_token", "Schema should contain 'api_token' attribute")
	assert.Contains(t, attrs, "base_url", "Schema should contain 'base_url' attribute")

	// Verify api_token is marked as sensitive
	if apiTokenAttr, ok := attrs["api_token"].(schema.StringAttribute); ok {
		assert.True(t, apiTokenAttr.Sensitive, "api_token should be marked as sensitive")
	} else {
		t.Error("api_token should be a StringAttribute")
	}
}

// TestProviderConfigure tests provider configuration logic
func TestProviderConfigure(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		envVars     map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration with explicit values",
			config: `
				provider "cachefly" {
					api_token = "test-token"
					base_url  = "https://api.test.cachefly.com/api/2.5"
				}
			`,
			expectError: false,
		},
		{
			name: "configuration with environment variables",
			config: `
				provider "cachefly" {}
			`,
			envVars: map[string]string{
				"CACHEFLY_API_TOKEN": "env-token",
				"CACHEFLY_BASE_URL":  "https://api.env.cachefly.com/api/2.5",
			},
			expectError: false,
		},
		{
			name: "missing api_token should fail",
			config: `
				provider "cachefly" {
					base_url = "https://api.test.cachefly.com/api/2.5"
				}
			`,
			expectError: true,
			errorMsg:    "Missing API Token",
		},
		{
			name: "default base_url should be used",
			config: `
				provider "cachefly" {
					api_token = "test-token"
				}
			`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for this test
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			if tt.expectError {
				resource.Test(t, resource.TestCase{
					ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						{
							Config:      tt.config,
							ExpectError: regexp.MustCompile(tt.errorMsg),
						},
					},
				})
			} else {
				resource.Test(t, resource.TestCase{
					ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						{
							Config: tt.config,
						},
					},
				})
			}
		})
	}
}

// TestGetConfigValue tests the helper function
func TestGetConfigValue(t *testing.T) {
	tests := []struct {
		name         string
		configValue  string
		envVar       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "config value takes precedence",
			configValue:  "config-value",
			envVar:       "TEST_VAR",
			envValue:     "env-value",
			defaultValue: "default-value",
			expected:     "config-value",
		},
		{
			name:         "env value used when config is empty",
			configValue:  "",
			envVar:       "TEST_VAR",
			envValue:     "env-value",
			defaultValue: "default-value",
			expected:     "env-value",
		},
		{
			name:         "default value used when both config and env are empty",
			configValue:  "",
			envVar:       "TEST_VAR",
			envValue:     "",
			defaultValue: "default-value",
			expected:     "default-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			}

		})
	}
}

// testAccPreCheck validates that required environment variables are set for acceptance tests
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	if v := os.Getenv("CACHEFLY_API_TOKEN"); v == "" {
		t.Fatal("CACHEFLY_API_TOKEN must be set for acceptance tests")
	}
}

// Test that all expected resources are registered
func TestProviderResources(t *testing.T) {
	ctx := context.Background()
	provider := &CacheFlyProvider{version: "test"}

	resources := provider.Resources(ctx)

	expectedResourceCount := 7 // we will update this based on our provider
	assert.Len(t, resources, expectedResourceCount, "Should have expected number of resources")

	// Test that each resource can be instantiated
	for i, resourceFunc := range resources {
		resource := resourceFunc()
		assert.NotNil(t, resource, "Resource %d should not be nil", i)
	}
}

// Test that all expected data sources are registered
func TestProviderDataSources(t *testing.T) {
	ctx := context.Background()
	provider := &CacheFlyProvider{version: "test"}

	dataSources := provider.DataSources(ctx)

	expectedDataSourceCount := 7 //
	assert.Len(t, dataSources, expectedDataSourceCount, "Should have expected number of data sources")

	// Test that each data source can be instantiated
	for i, dataSourceFunc := range dataSources {
		dataSource := dataSourceFunc()
		assert.NotNil(t, dataSource, "Data source %d should not be nil", i)
	}
}
