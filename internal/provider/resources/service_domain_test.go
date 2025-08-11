package resources_test

import (
	"context"
	"fmt"
	"strings"
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
func TestServiceDomainResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := resources.NewServiceDomainResource()

	req := fwresource.SchemaRequest{}
	resp := &fwresource.SchemaResponse{}
	r.Schema(ctx, req, resp)

	// no errors
	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	// required attributes exist
	attrs := resp.Schema.Attributes
	assert.Contains(t, attrs, "id")
	assert.Contains(t, attrs, "service_id")
	assert.Contains(t, attrs, "name")

	// optional attributes exist
	assert.Contains(t, attrs, "description")
	assert.Contains(t, attrs, "validation_mode")

	// computed attributes exist
	assert.Contains(t, attrs, "validation_target")
	assert.Contains(t, attrs, "validation_status")
	assert.Contains(t, attrs, "certificates")
	assert.Contains(t, attrs, "created_at")
	assert.Contains(t, attrs, "updated_at")
}

// Test Resource metadata
func TestServiceDomainResourceMetadata(t *testing.T) {
	ctx := context.Background()
	r := resources.NewServiceDomainResource()

	req := fwresource.MetadataRequest{
		ProviderTypeName: "cachefly",
	}
	resp := &fwresource.MetadataResponse{}
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "cachefly_service_domain", resp.TypeName)
}

// Test Configure error handling
func TestServiceDomainResourceConfigure(t *testing.T) {
	ctx := context.Background()
	r := resources.NewServiceDomainResource().(*resources.ServiceDomainResource)

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

// Test Import ID parsing logic
func TestServiceDomainResourceImportIDParsing(t *testing.T) {
	tests := []struct {
		name        string
		importID    string
		expectError bool
		serviceID   string
		domainID    string
	}{
		{
			name:        "valid import format",
			importID:    "service123:domain456",
			expectError: false,
			serviceID:   "service123",
			domainID:    "domain456",
		},
		{
			name:        "invalid import format - no colon",
			importID:    "service123domain456",
			expectError: true,
		},
		{
			name:        "invalid import format - too many parts",
			importID:    "service123:domain456:extra",
			expectError: true,
		},
		{
			name:        "invalid import format - empty",
			importID:    "",
			expectError: true,
		},
		{
			name:        "invalid import format - only colon",
			importID:    ":",
			expectError: true,
		},
		{
			name:        "invalid import format - empty service_id",
			importID:    ":domain456",
			expectError: true,
		},
		{
			name:        "invalid import format - empty domain_id",
			importID:    "service123:",
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			parts := strings.Split(tt.importID, ":")
			isValid := len(parts) == 2 && parts[0] != "" && parts[1] != ""

			if tt.expectError {
				assert.False(t, isValid, "Should be invalid for ID: %s", tt.importID)
			} else {
				assert.True(t, isValid, "Should be valid for ID: %s", tt.importID)
				if isValid {
					assert.Equal(t, tt.serviceID, parts[0], "Service ID should match")
					assert.Equal(t, tt.domainID, parts[1], "Domain ID should match")
				}
			}
		})
	}
}
func TestAccServiceDomainResource(t *testing.T) {
	rName := "test-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	domainName := rName + ".example.com"
	updatedDomainName := rName + "-updated.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkServiceDomainDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing with validation mode
			{
				Config: testAccServiceDomainResourceConfig(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceDomainExists("cachefly_service_domain."+rName),
					resource.TestCheckResourceAttr("cachefly_service_domain."+rName, "name", domainName),
					resource.TestCheckResourceAttr("cachefly_service_domain."+rName, "description", rName+" domain description"),
					resource.TestCheckResourceAttr("cachefly_service_domain."+rName, "validation_mode", "HTTP"),
					resource.TestCheckResourceAttrPair("cachefly_service_domain."+rName, "service_id", "cachefly_service."+rName, "id"),
					resource.TestCheckResourceAttrSet("cachefly_service_domain."+rName, "id"),
					resource.TestCheckResourceAttr("cachefly_service_domain."+rName, "validation_status", ""),
					resource.TestCheckResourceAttrSet("cachefly_service_domain."+rName, "created_at"),
					resource.TestCheckResourceAttrSet("cachefly_service_domain."+rName, "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "cachefly_service_domain." + rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccServiceDomainImportStateIdFunc("cachefly_service_domain." + rName),
			},
			// Update testing
			{
				Config: testAccServiceDomainResourceConfigUpdated(rName, updatedDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceDomainExists("cachefly_service_domain."+rName),
					resource.TestCheckResourceAttr("cachefly_service_domain."+rName, "name", updatedDomainName),
					resource.TestCheckResourceAttr("cachefly_service_domain."+rName, "description", "Updated domain description"),
					resource.TestCheckResourceAttr("cachefly_service_domain."+rName, "validation_mode", "DNS"),
				),
			},
		},
	})
}

// Helper function to check if service domain exists
func testAccCheckServiceDomainExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Service Domain ID is set")
		}
		serviceID := rs.Primary.Attributes["service_id"]
		if serviceID == "" {
			return fmt.Errorf("No Service ID is set")
		}
		sdkClient := provider.GetSDKClient()
		if sdkClient == nil {
			return fmt.Errorf("Failed to create CacheFly client")
		}
		// Check if the service domain exists via API call
		_, err := sdkClient.ServiceDomains.GetByID(context.Background(), serviceID, rs.Primary.ID, "")
		if err != nil {
			return fmt.Errorf("Service Domain %s not found: %s", rs.Primary.ID, err.Error())
		}
		return nil
	}
}

// Helper function to check if service domain is destroyed
func checkServiceDomainDestroy(s *terraform.State) error {
	sdkClient := provider.GetSDKClient()
	if sdkClient == nil {
		return fmt.Errorf("Failed to create CacheFly client")
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cachefly_service_domain" {
			continue
		}
		serviceID := rs.Primary.Attributes["service_id"]
		if serviceID == "" {
			continue
		}
		// Try to find the service domain
		_, err := sdkClient.ServiceDomains.GetByID(context.Background(), serviceID, rs.Primary.ID, "")

		if err != nil {
			// If we get an error, the domain likely doesn't exist, which is what we want
			// In a real scenario, you might want to check for specific error types
			return nil
		}
		return fmt.Errorf("Service Domain %s still exists", rs.Primary.ID)
	}
	return nil
}

// Helper function to generate import state ID for testing
func testAccServiceDomainImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		serviceID := rs.Primary.Attributes["service_id"]
		domainID := rs.Primary.ID

		if serviceID == "" || domainID == "" {
			return "", fmt.Errorf("Missing service_id or domain id")
		}
		return fmt.Sprintf("%s:%s", serviceID, domainID), nil
	}
}

// Test configuration for basic service domain
func testAccServiceDomainResourceConfig(name, domainName string) string {
	return fmt.Sprintf(`
provider "cachefly" {}
resource "cachefly_service" %[1]q {
  name        = %[1]q
  unique_name = "%[1]s-unique"
  description = "%[1]s service for domain testing"
}
resource "cachefly_service_domain" %[1]q {
  service_id      = cachefly_service.%[1]s.id
  name            = %[2]q
  description     = "%[1]s domain description"
  validation_mode = "HTTP"
}`, name, domainName)
}

// Test configuration for updated service domain
func testAccServiceDomainResourceConfigUpdated(name, domainName string) string {
	return fmt.Sprintf(`
provider "cachefly" {}
resource "cachefly_service" %[1]q {
  name        = %[1]q
  unique_name = "%[1]s-unique"
  description = "%[1]s service for domain testing"
}
resource "cachefly_service_domain" %[1]q {
  service_id      = cachefly_service.%[1]s.id
  name            = %[2]q
  description     = "Updated domain description"
  validation_mode = "DNS"
}`, name, domainName)
}
