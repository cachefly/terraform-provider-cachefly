package resources_test

import (
	"context"
	"fmt"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider"
	"github.com/cachefly/terraform-provider-cachefly/internal/provider/resources"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

// Test Schema validation
func TestLogTargetResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := resources.NewLogTargetResource()

	req := fwresource.SchemaRequest{}
	resp := &fwresource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	// Verify no errors
	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	// required attributes exist
	attrs := resp.Schema.Attributes
	assert.Contains(t, attrs, "id")
	assert.Contains(t, attrs, "name")
	assert.Contains(t, attrs, "type")

	// common log delivery options exist
	assert.Contains(t, attrs, "format")
	assert.Contains(t, attrs, "compression")
	assert.Contains(t, attrs, "sampling")

	// S3_BUCKET attributes exist
	assert.Contains(t, attrs, "endpoint")
	assert.Contains(t, attrs, "region")
	assert.Contains(t, attrs, "bucket")
	assert.Contains(t, attrs, "access_key")
	assert.Contains(t, attrs, "secret_key")
	assert.Contains(t, attrs, "signature_version")

	// GOOGLE_BUCKET attributes exist
	assert.Contains(t, attrs, "json_key")

	// AZURE_BLOB attributes exist
	assert.Contains(t, attrs, "endpoint_protocol")
	assert.Contains(t, attrs, "endpoint_suffix")
	assert.Contains(t, attrs, "account_name")
	assert.Contains(t, attrs, "account_key")
	assert.Contains(t, attrs, "container_name")
	assert.Contains(t, attrs, "prefix")

	// HTTP attributes exist
	assert.Contains(t, attrs, "uri")
	assert.Contains(t, attrs, "method")
	assert.Contains(t, attrs, "auth")
	assert.Contains(t, attrs, "username")
	assert.Contains(t, attrs, "password")
	assert.Contains(t, attrs, "token")

	// services logging attributes exist
	assert.Contains(t, attrs, "access_logs_services")
	assert.Contains(t, attrs, "origin_logs_services")

	// computed attributes exist
	assert.Contains(t, attrs, "created_at")
	assert.Contains(t, attrs, "updated_at")

	// removed Elasticsearch-era attributes are gone
	assert.NotContains(t, attrs, "hosts")
	assert.NotContains(t, attrs, "ssl")
	assert.NotContains(t, attrs, "ssl_certificate_verification")
	assert.NotContains(t, attrs, "index")
	assert.NotContains(t, attrs, "user")
	assert.NotContains(t, attrs, "api_key")

	// Verify sensitive attributes are marked as sensitive
	assert.True(t, attrs["access_key"].IsSensitive(), "access_key should be marked as sensitive")
	assert.True(t, attrs["secret_key"].IsSensitive(), "secret_key should be marked as sensitive")
	assert.True(t, attrs["json_key"].IsSensitive(), "json_key should be marked as sensitive")
	assert.True(t, attrs["account_key"].IsSensitive(), "account_key should be marked as sensitive")
	assert.True(t, attrs["password"].IsSensitive(), "password should be marked as sensitive")
	assert.True(t, attrs["token"].IsSensitive(), "token should be marked as sensitive")
}

// Test Resource metadata
func TestLogTargetResourceMetadata(t *testing.T) {
	ctx := context.Background()
	r := resources.NewLogTargetResource()

	req := fwresource.MetadataRequest{
		ProviderTypeName: "cachefly",
	}
	resp := &fwresource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "cachefly_log_target", resp.TypeName)
}

// Test Configure error handling
func TestLogTargetResourceConfigure(t *testing.T) {
	ctx := context.Background()
	r := resources.NewLogTargetResource().(*resources.LogTargetResource)

	// Test with nil provider data (should not error)
	req := fwresource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &fwresource.ConfigureResponse{}

	r.Configure(ctx, req, resp)
	assert.False(t, resp.Diagnostics.HasError(), "Should not error with nil provider data")

	// Test with wrong type should error
	req.ProviderData = "wrong-type"
	resp = &fwresource.ConfigureResponse{}

	r.Configure(ctx, req, resp)
	assert.True(t, resp.Diagnostics.HasError(), "Should error with wrong provider data type")
}

func TestAccLogTargetResourceS3(t *testing.T) {
	rName := "test-s3-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkLogTargetDestroy,
		Steps: []resource.TestStep{
			// Create S3 log target
			{
				Config: testAccLogTargetResourceConfigS3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogTargetExists("cachefly_log_target."+rName),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "name", rName),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "type", "S3_BUCKET"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "bucket", "my-log-bucket"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "region", "us-east-1"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "signature_version", "v4"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "format", "JSON"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "compression", "NONE"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "sampling", "100"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "access_logs_services.#", "1"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "origin_logs_services.#", "1"),
					resource.TestCheckResourceAttrPair("cachefly_log_target."+rName, "access_logs_services.0", "cachefly_service."+rName, "id"),
					resource.TestCheckResourceAttrPair("cachefly_log_target."+rName, "origin_logs_services.0", "cachefly_service."+rName, "id"),
					resource.TestCheckResourceAttrSet("cachefly_log_target."+rName, "id"),
					resource.TestCheckResourceAttrSet("cachefly_log_target."+rName, "created_at"),
					resource.TestCheckResourceAttrSet("cachefly_log_target."+rName, "updated_at"),
				),
			},
			// ImportState testing for S3 log target
			{
				ResourceName:      "cachefly_log_target." + rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// Secret key is sensitive and won't be returned in read operations
					"secret_key",
				},
			},
			// Update testing for S3 log target
			{
				Config: testAccLogTargetResourceConfigS3Updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogTargetExists("cachefly_log_target."+rName),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "name", rName+"-updated"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "type", "S3_BUCKET"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "bucket", "my-log-bucket-updated"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "region", "us-west-2"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "compression", "GZIP"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "sampling", "50"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "access_logs_services.#", "0"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "origin_logs_services.#", "0"),
				),
			},
		},
	})
}

func TestAccLogTargetResourceHTTP(t *testing.T) {
	rName := "test-http-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkLogTargetDestroy,
		Steps: []resource.TestStep{
			// Create HTTP log target
			{
				Config: testAccLogTargetResourceConfigHTTP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogTargetExists("cachefly_log_target."+rName),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "name", rName),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "type", "HTTP"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "uri", "https://logs.example.com/ingest"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "method", "POST"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "auth", "BASIC"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "username", "loguser"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "format", "NDJSON"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "compression", "GZIP"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "access_logs_services.#", "1"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "origin_logs_services.#", "1"),
					resource.TestCheckResourceAttrPair("cachefly_log_target."+rName, "access_logs_services.0", "cachefly_service."+rName, "id"),
					resource.TestCheckResourceAttrPair("cachefly_log_target."+rName, "origin_logs_services.0", "cachefly_service."+rName, "id"),
					resource.TestCheckResourceAttrSet("cachefly_log_target."+rName, "id"),
				),
			},
			// ImportState testing for HTTP log target
			{
				ResourceName:      "cachefly_log_target." + rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// Password is sensitive and won't be returned in read operations
					"password",
				},
			},
			// Update testing for HTTP log target
			{
				Config: testAccLogTargetResourceConfigHTTPUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogTargetExists("cachefly_log_target."+rName),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "name", rName+"-updated"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "type", "HTTP"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "uri", "https://logs-updated.example.com/ingest"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "method", "PUT"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "auth", "BEARER"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "access_logs_services.#", "0"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "origin_logs_services.#", "0"),
				),
			},
		},
	})
}

func TestAccLogTargetResourceAzureBlob(t *testing.T) {
	rName := "test-azure-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkLogTargetDestroy,
		Steps: []resource.TestStep{
			// Create Azure Blob log target
			{
				Config: testAccLogTargetResourceConfigAzureBlob(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogTargetExists("cachefly_log_target."+rName),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "name", rName),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "type", "AZURE_BLOB"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "account_name", "mystorageaccount"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "container_name", "cachefly-logs"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "prefix", "cdn/"),
					resource.TestCheckResourceAttrSet("cachefly_log_target."+rName, "id"),
					resource.TestCheckResourceAttrSet("cachefly_log_target."+rName, "created_at"),
					resource.TestCheckResourceAttrSet("cachefly_log_target."+rName, "updated_at"),
				),
			},
			// ImportState testing for Azure Blob log target
			{
				ResourceName:      "cachefly_log_target." + rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// Account key is sensitive and won't be returned in read operations
					"account_key",
				},
			},
			// Update testing for Azure Blob log target
			{
				Config: testAccLogTargetResourceConfigAzureBlobUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogTargetExists("cachefly_log_target."+rName),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "name", rName+"-updated"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "type", "AZURE_BLOB"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "container_name", "cachefly-logs-updated"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "prefix", "cdn-updated/"),
				),
			},
		},
	})
}

func TestAccLogTargetResourceGoogleCloud(t *testing.T) {
	rName := "test-gcp-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkLogTargetDestroy,
		Steps: []resource.TestStep{
			// Create Google Cloud log target
			{
				Config: testAccLogTargetResourceConfigGoogleCloud(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogTargetExists("cachefly_log_target."+rName),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "name", rName),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "type", "GOOGLE_BUCKET"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "bucket", "my-gcp-log-bucket"),
					resource.TestCheckResourceAttrSet("cachefly_log_target."+rName, "id"),
					resource.TestCheckResourceAttrSet("cachefly_log_target."+rName, "created_at"),
					resource.TestCheckResourceAttrSet("cachefly_log_target."+rName, "updated_at"),
				),
			},
			// ImportState testing for Google Cloud log target
			{
				ResourceName:      "cachefly_log_target." + rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// JSON key is sensitive and won't be returned in read operations
					"json_key",
				},
			},
			// Update testing for Google Cloud log target
			{
				Config: testAccLogTargetResourceConfigGoogleCloudUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogTargetExists("cachefly_log_target."+rName),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "name", rName+"-updated"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "type", "GOOGLE_BUCKET"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "bucket", "my-gcp-log-bucket-updated"),
				),
			},
		},
	})
}

// Helper function to check if log target exists
func testAccCheckLogTargetExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Log Target ID is set")
		}

		sdkClient := provider.GetSDKClient()
		if sdkClient == nil {
			return fmt.Errorf("Failed to create CacheFly client")
		}

		// Check if the log target exists via API call
		_, err := sdkClient.LogTargets.GetByID(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Log Target %s not found: %s", rs.Primary.ID, err.Error())
		}

		return nil
	}
}

// Helper function to check if log target is destroyed
func checkLogTargetDestroy(s *terraform.State) error {
	sdkClient := provider.GetSDKClient()
	if sdkClient == nil {
		return fmt.Errorf("Failed to create CacheFly client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cachefly_log_target" {
			continue
		}

		// Try to find the log target
		_, err := sdkClient.LogTargets.GetByID(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Log Target %s still exists", rs.Primary.ID)
		}

		// We expect an error indicating the log target doesn't exist
		// The exact error message may vary, but any error here indicates
		// the log target was successfully deleted
	}

	return nil
}

// Test configuration for S3 log target
func testAccLogTargetResourceConfigS3(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" %[1]q {
  name        = %[1]q
  unique_name = "%[1]s-unique"
  description = "%[1]s test service for log target"
}

resource "cachefly_log_target" %[1]q {
  name               = %[1]q
  type               = "S3_BUCKET"
  bucket             = "my-log-bucket"
  region             = "us-east-1"
  access_key         = "AKIAIOSFODNN7EXAMPLE"
  secret_key         = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  signature_version  = "v4"
  access_logs_services = [cachefly_service.%[1]s.id]
  origin_logs_services = [cachefly_service.%[1]s.id]

  depends_on = [cachefly_service.%[1]s]
}
`, name)
}

// Test configuration for UPDATED S3 log target
func testAccLogTargetResourceConfigS3Updated(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" %[1]q {
  name        = %[1]q
  unique_name = "%[1]s-unique"
  description = "%[1]s test service for log target"
}

resource "cachefly_log_target" %[1]q {
  name               = "%[1]s-updated"
  type               = "S3_BUCKET"
  bucket             = "my-log-bucket-updated"
  region             = "us-west-2"
  access_key         = "AKIAIOSFODNN7EXAMPLE"
  secret_key         = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  signature_version  = "v4"
  compression        = "GZIP"
  sampling           = 50
  access_logs_services = []
  origin_logs_services = []

  depends_on = [cachefly_service.%[1]s]
}
`, name)
}

// Test configuration for HTTP log target
func testAccLogTargetResourceConfigHTTP(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" %[1]q {
  name        = %[1]q
  unique_name = "%[1]s-unique"
  description = "%[1]s test service for log target"
}

resource "cachefly_log_target" %[1]q {
  name        = %[1]q
  type        = "HTTP"
  uri         = "https://logs.example.com/ingest"
  method      = "POST"
  auth        = "BASIC"
  username    = "loguser"
  password    = "secret-password"
  format      = "NDJSON"
  compression = "GZIP"
  access_logs_services = [cachefly_service.%[1]s.id]
  origin_logs_services = [cachefly_service.%[1]s.id]

  depends_on = [cachefly_service.%[1]s]
}
`, name)
}

// Test configuration for UPDATED HTTP log target
func testAccLogTargetResourceConfigHTTPUpdated(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" %[1]q {
  name        = %[1]q
  unique_name = "%[1]s-unique"
  description = "%[1]s test service for log target"
}

resource "cachefly_log_target" %[1]q {
  name        = "%[1]s-updated"
  type        = "HTTP"
  uri         = "https://logs-updated.example.com/ingest"
  method      = "PUT"
  auth        = "BEARER"
  token       = "secret-token"
  format      = "NDJSON"
  compression = "GZIP"
  access_logs_services = []
  origin_logs_services = []

  depends_on = [cachefly_service.%[1]s]
}
`, name)
}

// Test configuration for Azure Blob log target
func testAccLogTargetResourceConfigAzureBlob(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_log_target" %[1]q {
  name           = %[1]q
  type           = "AZURE_BLOB"
  account_name   = "mystorageaccount"
  account_key    = "bXktYWNjb3VudC1rZXktZXhhbXBsZQ=="
  container_name = "cachefly-logs"
  prefix         = "cdn/"
}
`, name)
}

// Test configuration for UPDATED Azure Blob log target
func testAccLogTargetResourceConfigAzureBlobUpdated(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_log_target" %[1]q {
  name           = "%[1]s-updated"
  type           = "AZURE_BLOB"
  account_name   = "mystorageaccount"
  account_key    = "bXktYWNjb3VudC1rZXktZXhhbXBsZQ=="
  container_name = "cachefly-logs-updated"
  prefix         = "cdn-updated/"
}
`, name)
}

// Test configuration for Google Cloud log target
func testAccLogTargetResourceConfigGoogleCloud(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_log_target" %[1]q {
  name      = %[1]q
  type      = "GOOGLE_BUCKET"
  bucket    = "my-gcp-log-bucket"
  json_key  = jsonencode({
    "type": "service_account",
    "project_id": "my-project-12345",
    "private_key_id": "key-id-12345",
    "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDExample...\n-----END PRIVATE KEY-----\n",
    "client_email": "my-service-account@my-project-12345.iam.gserviceaccount.com",
    "client_id": "123456789012345678901",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/my-service-account%%40my-project-12345.iam.gserviceaccount.com"
  })
}
`, name)
}

// Test configuration for UPDATED Google Cloud log target
func testAccLogTargetResourceConfigGoogleCloudUpdated(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_log_target" %[1]q {
  name      = "%[1]s-updated"
  type      = "GOOGLE_BUCKET"
  bucket    = "my-gcp-log-bucket-updated"
  json_key  = jsonencode({
    "type": "service_account",
    "project_id": "my-project-12345",
    "private_key_id": "key-id-12345",
    "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDExample...\n-----END PRIVATE KEY-----\n",
    "client_email": "my-service-account@my-project-12345.iam.gserviceaccount.com",
    "client_id": "123456789012345678901",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/my-service-account%%40my-project-12345.iam.gserviceaccount.com"
  })
}
`, name)
}
