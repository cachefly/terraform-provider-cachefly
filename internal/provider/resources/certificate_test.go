package resources_test

import (
	"context"
	"fmt"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"os"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider"
	"github.com/cachefly/terraform-provider-cachefly/internal/provider/resources"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

// Test 1: Schema validation
func TestCertificateResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := resources.NewCertificateResource()
	req := fwresource.SchemaRequest{}
	resp := &fwresource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	// Verify no errors
	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	attrs := resp.Schema.Attributes
	assert.Contains(t, attrs, "certificate")
	assert.Contains(t, attrs, "certificate_key")

	assert.Contains(t, attrs, "password")

	assert.Contains(t, attrs, "id")
	assert.Contains(t, attrs, "subject_common_name")
	assert.Contains(t, attrs, "subject_names")
	assert.Contains(t, attrs, "expired")
	assert.Contains(t, attrs, "expiring")
	assert.Contains(t, attrs, "in_use")
	assert.Contains(t, attrs, "managed")
	assert.Contains(t, attrs, "services")
	assert.Contains(t, attrs, "domains")
	assert.Contains(t, attrs, "not_before")
	assert.Contains(t, attrs, "not_after")
	assert.Contains(t, attrs, "created_at")

	assert.True(t, attrs["certificate"].IsRequired())
	assert.True(t, attrs["certificate_key"].IsRequired())

	assert.True(t, attrs["password"].IsOptional())

	assert.True(t, attrs["id"].IsComputed())
	assert.True(t, attrs["subject_common_name"].IsComputed())
	assert.True(t, attrs["expired"].IsComputed())

	assert.True(t, attrs["certificate"].IsSensitive())
	assert.True(t, attrs["certificate_key"].IsSensitive())
	assert.True(t, attrs["password"].IsSensitive())
}

// Test 2: Resource metadata
func TestCertificateResourceMetadata(t *testing.T) {
	ctx := context.Background()
	r := resources.NewCertificateResource()
	req := fwresource.MetadataRequest{
		ProviderTypeName: "cachefly",
	}
	resp := &fwresource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "cachefly_certificate", resp.TypeName)
}

// Test 3: Configure error handling
func TestCertificateResourceConfigure(t *testing.T) {
	ctx := context.Background()
	r := resources.NewCertificateResource().(*resources.CertificateResource)

	// Test with nil provider data
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

// Acceptance test: Create, Read, Import, and Destroy for certificate resource
func TestAccCertificateResource(t *testing.T) {
	rName := "test-cert-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t); certificateTestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateResourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists("cachefly_certificate."+rName),
					resource.TestCheckResourceAttrSet("cachefly_certificate."+rName, "id"),
					resource.TestCheckResourceAttrSet("cachefly_certificate."+rName, "created_at"),
					resource.TestCheckResourceAttrSet("cachefly_certificate."+rName, "subject_common_name"),
					resource.TestCheckResourceAttrSet("cachefly_certificate."+rName, "not_before"),
					resource.TestCheckResourceAttrSet("cachefly_certificate."+rName, "not_after"),
				),
			},
			{
				ResourceName:            "cachefly_certificate." + rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate", "certificate_key", "password"},
			},
		},
	})
}

// Helper: ensure required env vars for certificate tests are set
func certificateTestAccPreCheck(t *testing.T) {
	if os.Getenv("CF_TEST_CERTIFICATE") == "" || os.Getenv("CF_TEST_CERTIFICATE_KEY") == "" {
		t.Skip("Acceptance test skipped: CF_TEST_CERTIFICATE and CF_TEST_CERTIFICATE_KEY must be set to run certificate tests")
	}
}

// Helper: build test configuration for certificate resource
func testAccCertificateResourceConfig(name string) string {
	cert := os.Getenv("CF_TEST_CERTIFICATE")
	key := os.Getenv("CF_TEST_CERTIFICATE_KEY")

	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_certificate" %q {
  certificate = <<-EOT
%s
EOT

  certificate_key = <<-EOT
%s
EOT
}
`, name, cert, key)
}

// Helper: check certificate exists via API
func testAccCheckCertificateExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Certificate ID is set")
		}

		sdkClient := provider.GetSDKClient()
		if sdkClient == nil {
			return fmt.Errorf("Failed to create CacheFly client")
		}

		_, err := sdkClient.Certificates.GetByID(context.Background(), rs.Primary.ID, "")
		if err != nil {
			return fmt.Errorf("Certificate %s not found: %s", rs.Primary.ID, err.Error())
		}

		return nil
	}
}

// Helper: verify certificate is destroyed
func checkCertificateDestroy(s *terraform.State) error {
	sdkClient := provider.GetSDKClient()
	if sdkClient == nil {
		return fmt.Errorf("Failed to create CacheFly client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cachefly_certificate" {
			continue
		}

		_, err := sdkClient.Certificates.GetByID(context.Background(), rs.Primary.ID, "")
		if err == nil {
			return fmt.Errorf("Certificate %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
