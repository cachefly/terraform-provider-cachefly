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

	// optional attributes exist
	assert.Contains(t, attrs, "endpoint")
	assert.Contains(t, attrs, "region")
	assert.Contains(t, attrs, "bucket")
	assert.Contains(t, attrs, "access_key")
	assert.Contains(t, attrs, "secret_key")
	assert.Contains(t, attrs, "signature_version")
	assert.Contains(t, attrs, "json_key")
	assert.Contains(t, attrs, "hosts")
	assert.Contains(t, attrs, "ssl")
	assert.Contains(t, attrs, "ssl_certificate_verification")
	assert.Contains(t, attrs, "index")
	assert.Contains(t, attrs, "user")
	assert.Contains(t, attrs, "password")
	assert.Contains(t, attrs, "api_key")
	assert.Contains(t, attrs, "accessLogsServices")
	assert.Contains(t, attrs, "originLogsServices")

	// computed attributes exist
	assert.Contains(t, attrs, "created_at")
	assert.Contains(t, attrs, "updated_at")

	// Verify sensitive attributes are marked as sensitive
	assert.True(t, attrs["access_key"].IsSensitive(), "access_key should be marked as sensitive")
	assert.True(t, attrs["secret_key"].IsSensitive(), "secret_key should be marked as sensitive")
	assert.True(t, attrs["json_key"].IsSensitive(), "json_key should be marked as sensitive")
	assert.True(t, attrs["password"].IsSensitive(), "password should be marked as sensitive")
	assert.True(t, attrs["api_key"].IsSensitive(), "api_key should be marked as sensitive")
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
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "access_logs_services.#", "0"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "origin_logs_services.#", "0"),
				),
			},
		},
	})
}

func TestAccLogTargetResourceElasticsearch(t *testing.T) {
	rName := "test-es-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkLogTargetDestroy,
		Steps: []resource.TestStep{
			// Create Elasticsearch log target
			{
				Config: testAccLogTargetResourceConfigElasticsearch(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogTargetExists("cachefly_log_target."+rName),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "name", rName),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "type", "ELASTICSEARCH"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "hosts.#", "2"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "hosts.0", "elasticsearch1.example.com:9200"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "hosts.1", "elasticsearch2.example.com:9200"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "ssl", "true"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "ssl_certificate_verification", "true"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "index", "cachefly-logs"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "user", "elastic"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "access_logs_services.#", "1"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "origin_logs_services.#", "1"),
					resource.TestCheckResourceAttrPair("cachefly_log_target."+rName, "access_logs_services.0", "cachefly_service."+rName, "id"),
					resource.TestCheckResourceAttrPair("cachefly_log_target."+rName, "origin_logs_services.0", "cachefly_service."+rName, "id"),
					resource.TestCheckResourceAttrSet("cachefly_log_target."+rName, "id"),
				),
			},
			// ImportState testing for Elasticsearch log target
			{
				ResourceName:      "cachefly_log_target." + rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// Password is sensitive and won't be returned in read operations
					"password",
				},
			},
			// Update testing for Elasticsearch log target
			{
				Config: testAccLogTargetResourceConfigElasticsearchUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogTargetExists("cachefly_log_target."+rName),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "name", rName+"-updated"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "type", "ELASTICSEARCH"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "hosts.#", "2"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "hosts.0", "elasticsearch3.example.com:9200"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "hosts.1", "elasticsearch4.example.com:9200"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "ssl", "false"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "ssl_certificate_verification", "false"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "index", "cachefly-logs-updated"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "access_logs_services.#", "0"),
					resource.TestCheckResourceAttr("cachefly_log_target."+rName, "origin_logs_services.#", "0"),
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

		fmt.Printf("rs.Primary.ID: %s\n", rs.Primary.ID)
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

// Test configuration for basic syslog log target
func testAccLogTargetResourceConfig(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_log_target" %[1]q {
  name                           = %[1]q
  type                           = "SYSLOG"
  endpoint                       = "syslog.example.com:514"
  ssl                            = false
  ssl_certificate_verification   = true
}
`, name)
}

// Test configuration for updated syslog log target
func testAccLogTargetResourceConfigUpdated(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_log_target" %[1]q {
  name                           = "%[1]s-updated"
  type                           = "SYSLOG"
  endpoint                       = "updated-syslog.example.com:514"
  ssl                            = true
  ssl_certificate_verification   = false
}
`, name)
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
  access_logs_services = []
  origin_logs_services = []

  depends_on = [cachefly_service.%[1]s]
}
`, name)
}

// Test configuration for Elasticsearch log target
func testAccLogTargetResourceConfigElasticsearch(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" %[1]q {
  name        = %[1]q
  unique_name = "%[1]s-unique"
  description = "%[1]s test service for log target"
}

resource "cachefly_log_target" %[1]q {
  name                           = %[1]q
  type                           = "ELASTICSEARCH"
  hosts                          = [
    "elasticsearch1.example.com:9200",
    "elasticsearch2.example.com:9200"
  ]
  ssl                            = true
  ssl_certificate_verification   = true
  index                          = "cachefly-logs"
  user                           = "elastic"
  password                       = "secret-password"
  access_logs_services           = [cachefly_service.%[1]s.id]
  origin_logs_services           = [cachefly_service.%[1]s.id]

  depends_on = [cachefly_service.%[1]s]
}
`, name)
}

// Test configuration for UPDATED Elasticsearch log target
func testAccLogTargetResourceConfigElasticsearchUpdated(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" %[1]q {
  name        = %[1]q
  unique_name = "%[1]s-unique"
  description = "%[1]s test service for log target"
}

resource "cachefly_log_target" %[1]q {
  name                           = "%[1]s-updated"
  type                           = "ELASTICSEARCH"
  hosts                          = [
    "elasticsearch3.example.com:9200",
    "elasticsearch4.example.com:9200"
  ]
  ssl                            = false
  ssl_certificate_verification   = false
  index                          = "cachefly-logs-updated"
  user                           = "elastic"
  password                       = "secret-password"
  access_logs_services           = []
  origin_logs_services           = []

  depends_on = [cachefly_service.%[1]s]
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
