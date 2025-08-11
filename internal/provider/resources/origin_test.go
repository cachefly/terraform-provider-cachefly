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
func TestOriginResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := resources.NewOriginResource()

	req := fwresource.SchemaRequest{}
	resp := &fwresource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	// Verify no errors
	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	// required attributes exist
	attrs := resp.Schema.Attributes
	assert.Contains(t, attrs, "id")
	assert.Contains(t, attrs, "type")
	assert.Contains(t, attrs, "hostname")

	// optional attributes exist
	assert.Contains(t, attrs, "name")
	assert.Contains(t, attrs, "scheme")
	assert.Contains(t, attrs, "cache_by_query_param")
	assert.Contains(t, attrs, "gzip")
	assert.Contains(t, attrs, "ttl")
	assert.Contains(t, attrs, "missed_ttl")
	assert.Contains(t, attrs, "connection_timeout")
	assert.Contains(t, attrs, "time_to_first_byte_timeout")

	// S3-specific attributes exist
	assert.Contains(t, attrs, "access_key")
	assert.Contains(t, attrs, "secret_key")
	assert.Contains(t, attrs, "region")
	assert.Contains(t, attrs, "signature_version")

	// computed attributes exist
	assert.Contains(t, attrs, "created_at")
	assert.Contains(t, attrs, "updated_at")

	// Verify sensitive attributes are marked as sensitive
	assert.True(t, attrs["access_key"].IsSensitive(), "access_key should be marked as sensitive")
	assert.True(t, attrs["secret_key"].IsSensitive(), "secret_key should be marked as sensitive")
}

// Test Resource metadata
func TestOriginResourceMetadata(t *testing.T) {
	ctx := context.Background()
	r := resources.NewOriginResource()

	req := fwresource.MetadataRequest{
		ProviderTypeName: "cachefly",
	}
	resp := &fwresource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "cachefly_origin", resp.TypeName)
}

// Test Configure error handling
func TestOriginResourceConfigure(t *testing.T) {
	ctx := context.Background()
	r := resources.NewOriginResource().(*resources.OriginResource)

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

func TestAccOriginResource(t *testing.T) {
	rName := "test-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkOriginDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccOriginResourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOriginExists("cachefly_origin."+rName),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "name", rName),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "type", "WEB"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "hostname", "example.com"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "scheme", "HTTPS"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "cache_by_query_param", "false"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "gzip", "true"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "ttl", "86400"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "missed_ttl", "300"),
					resource.TestCheckResourceAttrSet("cachefly_origin."+rName, "id"),
					resource.TestCheckResourceAttrSet("cachefly_origin."+rName, "created_at"),
					resource.TestCheckResourceAttrSet("cachefly_origin."+rName, "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "cachefly_origin." + rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing
			{
				Config: testAccOriginResourceConfigUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOriginExists("cachefly_origin."+rName),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "name", rName+"-updated"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "type", "WEB"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "hostname", "updated.example.com"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "scheme", "HTTP"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "cache_by_query_param", "true"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "gzip", "false"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "ttl", "3600"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "missed_ttl", "600"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "connection_timeout", "10"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "time_to_first_byte_timeout", "10"),
				),
			},
		},
	})
}

func TestAccOriginResourceS3(t *testing.T) {
	rName := "test-s3-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkOriginDestroy,
		Steps: []resource.TestStep{
			// Create S3 origin
			{
				Config: testAccOriginResourceConfigS3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOriginExists("cachefly_origin."+rName),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "name", rName),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "type", "S3_BUCKET"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "hostname", "my-bucket.s3.amazonaws.com"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "access_key", "AKIAIOSFODNN7EXAMPLE"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "region", "us-east-1"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "signature_version", "v4"),
					resource.TestCheckResourceAttrSet("cachefly_origin."+rName, "id"),
					resource.TestCheckResourceAttrSet("cachefly_origin."+rName, "created_at"),
					resource.TestCheckResourceAttrSet("cachefly_origin."+rName, "updated_at"),
				),
			},
			// ImportState testing for S3 origin
			{
				ResourceName:      "cachefly_origin." + rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// Secret key is sensitive and won't be returned in read operations
					"secret_key",
				},
			},
		},
	})
}

func TestAccOriginResourceMinimal(t *testing.T) {
	rName := "test-minimal-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkOriginDestroy,
		Steps: []resource.TestStep{
			// Create minimal origin with only required fields
			{
				Config: testAccOriginResourceConfigMinimal(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOriginExists("cachefly_origin."+rName),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "type", "WEB"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "hostname", "minimal.example.com"),
					// Check computed defaults
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "scheme", "FOLLOW"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "connection_timeout", "3"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "cache_by_query_param", "false"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "gzip", "false"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "ttl", "2678400"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "missed_ttl", "86400"),
					resource.TestCheckResourceAttr("cachefly_origin."+rName, "time_to_first_byte_timeout", "3"),
					resource.TestCheckResourceAttrSet("cachefly_origin."+rName, "id"),
				),
			},
		},
	})
}

// Helper function to check if origin exists
func testAccCheckOriginExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Origin ID is set")
		}

		sdkClient := provider.GetSDKClient()
		if sdkClient == nil {
			return fmt.Errorf("Failed to create CacheFly client")
		}

		// Check if the origin exists via API call
		_, err := sdkClient.Origins.GetByID(context.Background(), rs.Primary.ID, "")
		if err != nil {
			return fmt.Errorf("Origin %s not found: %s", rs.Primary.ID, err.Error())
		}

		return nil
	}
}

// Helper function to check if origin is destroyed
func checkOriginDestroy(s *terraform.State) error {
	sdkClient := provider.GetSDKClient()
	if sdkClient == nil {
		return fmt.Errorf("Failed to create CacheFly client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cachefly_origin" {
			continue
		}

		// Try to find the origin
		_, err := sdkClient.Origins.GetByID(context.Background(), rs.Primary.ID, "")
		if err == nil {
			return fmt.Errorf("Origin %s still exists", rs.Primary.ID)
		}

		// We expect an error indicating the origin doesn't exist
		// The exact error message may vary, but any error here indicates
		// the origin was successfully deleted
	}

	return nil
}

// Test configuration for basic origin
func testAccOriginResourceConfig(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_origin" %[1]q {
  name                   = %[1]q
  type                   = "WEB"
  hostname               = "example.com"
  scheme                 = "HTTPS"
  cache_by_query_param   = false
  gzip                   = true
  ttl                    = 86400
  missed_ttl             = 300
}
`, name)
}

// Test configuration for updated origin
func testAccOriginResourceConfigUpdated(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_origin" %[1]q {
  name                          = "%[1]s-updated"
  type                          = "WEB"
  hostname                          = "updated.example.com"
  scheme                        = "HTTP"
  cache_by_query_param          = true
  gzip                          = false
  ttl                           = 3600
  missed_ttl                    = 600
  connection_timeout            = 10
  time_to_first_byte_timeout    = 10
}
`, name)
}

// Test configuration for S3 origin
func testAccOriginResourceConfigS3(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_origin" %[1]q {
  name              = %[1]q
  type              = "S3_BUCKET"
  hostname          = "my-bucket.s3.amazonaws.com"
  access_key        = "AKIAIOSFODNN7EXAMPLE"
  secret_key        = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  region            = "us-east-1"
  signature_version = "v4"
}
`, name)
}

// Test configuration for minimal origin (only required fields)
func testAccOriginResourceConfigMinimal(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_origin" %[1]q {
  type = "WEB"
  hostname = "minimal.example.com"
}
`, name)
}
