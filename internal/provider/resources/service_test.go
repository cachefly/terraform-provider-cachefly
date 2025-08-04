package resources_test

import (
	"context"
	"fmt"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"
	"github.com/cachefly/terraform-provider-cachefly/internal/provider"
	"github.com/cachefly/terraform-provider-cachefly/internal/provider/resources"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

// Test Schema validation
func TestServiceResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := resources.NewServiceResource()

	req := fwresource.SchemaRequest{}
	resp := &fwresource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	// Verify no errors
	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	// required attributes exist
	attrs := resp.Schema.Attributes
	assert.Contains(t, attrs, "id")
	assert.Contains(t, attrs, "name")
	assert.Contains(t, attrs, "unique_name")

	// optional attributes exist
	assert.Contains(t, attrs, "description")
	assert.Contains(t, attrs, "auto_ssl")
	assert.Contains(t, attrs, "configuration_mode")

	// computed attributes exist
	assert.Contains(t, attrs, "status")
	assert.Contains(t, attrs, "created_at")
	assert.Contains(t, attrs, "updated_at")
}

// Test Resource metadata
func TestServiceResourceMetadata(t *testing.T) {
	ctx := context.Background()
	r := resources.NewServiceResource()

	req := fwresource.MetadataRequest{
		ProviderTypeName: "cachefly",
	}
	resp := &fwresource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "cachefly_service", resp.TypeName)
}

// Test Configure error handling
func TestServiceResourceConfigure(t *testing.T) {
	ctx := context.Background()
	r := resources.NewServiceResource().(*resources.ServiceResource)

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

func TestAccServiceResource(t *testing.T) {
	sdkClient := provider.GetSDKClient()
	rName := "test-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	tlsProfiles, err := sdkClient.TLSProfiles.List(context.Background(), v2_5.ListTLSProfilesOptions{})
	if err != nil {
		t.Fatalf("Failed to get TLS profiles: %v", err)
	}
	tlsProfileId := tlsProfiles.Profiles[0].ID

	deliveryRegions, err := sdkClient.DeliveryRegions.List(context.Background(), v2_5.ListDeliveryRegionsOptions{})
	if err != nil {
		t.Fatalf("Failed to get services: %v", err)
	}
	deliveryRegionId := deliveryRegions.Regions[0].ID

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkServiceDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccServiceResourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceExists("cachefly_service."+rName),
					resource.TestCheckResourceAttr("cachefly_service."+rName, "name", rName),
					resource.TestCheckResourceAttr("cachefly_service."+rName, "unique_name", rName+"-unique"),
					resource.TestCheckResourceAttr("cachefly_service."+rName, "description", rName+" description"),
					resource.TestCheckResourceAttrSet("cachefly_service."+rName, "id"),
					resource.TestCheckResourceAttrSet("cachefly_service."+rName, "status"),
					resource.TestCheckResourceAttrSet("cachefly_service."+rName, "created_at"),
					resource.TestCheckResourceAttrSet("cachefly_service."+rName, "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "cachefly_service." + rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing
			{
				Config: testAccServiceResourceConfigWithUpdatedFields(rName, tlsProfileId, deliveryRegionId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceExists("cachefly_service."+rName),
					resource.TestCheckResourceAttr("cachefly_service."+rName, "name", rName),
					resource.TestCheckResourceAttr("cachefly_service."+rName, "unique_name", rName+"-unique"),
					resource.TestCheckResourceAttr("cachefly_service."+rName, "description", "Updated service description"),
					resource.TestCheckResourceAttr("cachefly_service."+rName, "tls_profile", tlsProfileId),
					resource.TestCheckResourceAttr("cachefly_service."+rName, "delivery_region", deliveryRegionId),
				),
			},
		},
	})
}

func TestAccServiceResourceWithOptions(t *testing.T) {
	rName := "test-options-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "cachefly_service." + rName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkServiceDestroy,
		Steps: []resource.TestStep{
			// Create service with options
			{
				Config: testAccServiceResourceConfigWithOptions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "unique_name", rName+"-unique"),
					resource.TestCheckResourceAttr(resourceName, "options.%", "4"),
				),
			},
			// Update service options (test differential updates)
			{
				Config: testAccServiceResourceConfigWithUpdatedOptions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					// Check updated option values
					resource.TestCheckResourceAttr(resourceName, "options.autoRedirect", "false"),
					resource.TestCheckResourceAttr(resourceName, "options.protectServeKeyEnabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "options.purgemode.enabled", "false"),
				),
			},
			// ImportState testing with options
			{
				ResourceName: "cachefly_service." + rName,
				ImportState:  true,
				// ImportStateVerify: true,
			},
		},
	})
}

// Test import scenario where all options should be loaded
func TestAccServiceResourceImportWithOptions(t *testing.T) {
	rName := "test-import-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkServiceDestroy,
		Steps: []resource.TestStep{
			// Create service with options outside of Terraform (simulated)
			{
				Config: testAccServiceResourceConfigWithOptions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceExists("cachefly_service." + rName),
				),
			},
			// Import the service - should load all available options
			{
				ResourceName:      "cachefly_service." + rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// Ignore computed fields that might differ slightly
					"updated_at",
				},
			},
		},
	})
}

// Helper function to check if service exists
func testAccCheckServiceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Service ID is set")
		}

		fmt.Printf("rs.Primary.ID: %s\n", rs.Primary.ID)
		sdkClient := provider.GetSDKClient()
		if sdkClient == nil {
			return fmt.Errorf("Failed to create CacheFly client")
		}

		// Check if the service exists via API call
		_, err := sdkClient.Services.GetByID(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Service %s not found: %s", rs.Primary.ID, err.Error())
		}

		return nil
	}
}

// Helper function to check if service is destroyed
func checkServiceDestroy(s *terraform.State) error {
	sdkClient := provider.GetSDKClient()
	if sdkClient == nil {
		return fmt.Errorf("Failed to create CacheFly client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cachefly_service" {
			continue
		}

		// For CacheFly, services are deactivated rather than deleted, so we should
		// check that the error indicates the service is deactivated
		service, err := sdkClient.Services.GetByID(context.Background(), rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("API error when checking if service %s exists: %s", rs.Primary.ID, err.Error())
		}

		if service.Status == "DEACTIVATED" {
			return nil
		} else {
			return fmt.Errorf("Service %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

// Test configuration for basic service
func testAccServiceResourceConfig(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" %[1]q {
  name        = %[1]q
  unique_name = "%[1]s-unique"
  description = "%[1]s description"
}
`, name)
}

// Test configuration for service with optional fields
func testAccServiceResourceConfigWithUpdatedFields(name string, tlsProfileId string, deliveryRegionId string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" %[1]q {
  name               = %[1]q
  unique_name        = "%[1]s-unique"
  description        = "Updated service description"
  tls_profile        = %[2]q
  delivery_region    = %[3]q
}
`, name, tlsProfileId, deliveryRegionId)
}

// Test configuration for service with basic options
func testAccServiceResourceConfigWithOptions(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" %[1]q {
  name        = %[1]q
  unique_name = "%[1]s-unique"
  description = "%[1]s description"
  
  options = {
    autoRedirect = true
	reverseProxy = {
		enabled = true
        originScheme = "HTTP"
        cacheByQueryParam = true
        useRobotsTxt = true
        ttl = 123
		hostname = "abc.com"
	}
	protectServeKeyEnabled = true
	purgemode = {
        enabled = true
        value = {
            exact = true
            directory = true
            extension = true
        }
    }
  }
}
`, name)
}

// Test configuration for service with updated options
func testAccServiceResourceConfigWithUpdatedOptions(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" %[1]q {
  name        = %[1]q
  unique_name = "%[1]s-unique"
  description = "%[1]s description"
  
  options = {
	autoRedirect = false
	purgemode = {
		enabled = false
	}
	protectServeKeyEnabled = false
  }
}
`, name)
}
