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
func TestUserResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := resources.NewUserResource()

	req := fwresource.SchemaRequest{}
	resp := &fwresource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	// Verify no errors
	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	// required attributes exist
	attrs := resp.Schema.Attributes
	assert.Contains(t, attrs, "id")
	assert.Contains(t, attrs, "username")
	assert.Contains(t, attrs, "email")
	assert.Contains(t, attrs, "password")
	assert.Contains(t, attrs, "full_name")

	// optional attributes exist
	assert.Contains(t, attrs, "phone")
	assert.Contains(t, attrs, "password_change_required")
	assert.Contains(t, attrs, "services")
	assert.Contains(t, attrs, "permissions")

	// computed attributes exist
	assert.Contains(t, attrs, "status")
	assert.Contains(t, attrs, "created_at")
	assert.Contains(t, attrs, "updated_at")

	// Verify password is sensitive
	assert.True(t, attrs["password"].IsSensitive())
}

// Test Resource metadata
func TestUserResourceMetadata(t *testing.T) {
	ctx := context.Background()
	r := resources.NewUserResource()

	req := fwresource.MetadataRequest{
		ProviderTypeName: "cachefly",
	}
	resp := &fwresource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "cachefly_user", resp.TypeName)
}

// Test Configure error handling
func TestUserResourceConfigure(t *testing.T) {
	ctx := context.Background()
	r := resources.NewUserResource().(*resources.UserResource)

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

func TestAccUserResource(t *testing.T) {
	rName := "test-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	email := rName + "@example.com"
	updatedEmail := rName + "-updated@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkUserDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserResourceConfig(rName, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserExists("cachefly_user."+rName),
					resource.TestCheckResourceAttr("cachefly_user."+rName, "username", rName),
					resource.TestCheckResourceAttr("cachefly_user."+rName, "email", email),
					resource.TestCheckResourceAttr("cachefly_user."+rName, "full_name", rName+" User"),
					resource.TestCheckResourceAttr("cachefly_user."+rName, "phone", "+1234567890"),
					resource.TestCheckResourceAttr("cachefly_user."+rName, "password_change_required", "false"),
					resource.TestCheckResourceAttrSet("cachefly_user."+rName, "id"),
					resource.TestCheckResourceAttrSet("cachefly_user."+rName, "status"),
					resource.TestCheckResourceAttrSet("cachefly_user."+rName, "created_at"),
					resource.TestCheckResourceAttrSet("cachefly_user."+rName, "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "cachefly_user." + rName,
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore password since it's sensitive and not returned by read
				ImportStateVerifyIgnore: []string{"password"},
			},
			// Update testing
			{
				Config: testAccUserResourceConfigUpdated(rName, updatedEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserExists("cachefly_user."+rName),
					resource.TestCheckResourceAttr("cachefly_user."+rName, "username", rName),
					resource.TestCheckResourceAttr("cachefly_user."+rName, "email", updatedEmail),
					resource.TestCheckResourceAttr("cachefly_user."+rName, "full_name", rName+" Updated User"),
					resource.TestCheckResourceAttr("cachefly_user."+rName, "phone", "+0987654321"),
					resource.TestCheckResourceAttr("cachefly_user."+rName, "password_change_required", "true"),
				),
			},
		},
	})
}

func TestAccUserResourceWithServicesAndPermissions(t *testing.T) {
	rName := "test-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	email := rName + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkUserDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing with services and permissions
			{
				Config: testAccUserResourceConfigWithServicesAndPermissions(rName, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserExists("cachefly_user."+rName),
					resource.TestCheckResourceAttr("cachefly_user."+rName, "username", rName),
					resource.TestCheckResourceAttr("cachefly_user."+rName, "email", email),
					resource.TestCheckResourceAttr("cachefly_user."+rName, "full_name", rName+" Service User"),
					resource.TestCheckResourceAttr("cachefly_user."+rName, "services.#", "1"),
					resource.TestCheckResourceAttrPair("cachefly_user."+rName, "services.0", "cachefly_service."+rName+"-svc", "id"),
					resource.TestCheckResourceAttr("cachefly_user."+rName, "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("cachefly_user."+rName, "permissions.*", "P_USER_PROFILE_VIEW"),
					resource.TestCheckTypeSetElemAttr("cachefly_user."+rName, "permissions.*", "P_SERVICE_ALL"),
					resource.TestCheckResourceAttrSet("cachefly_user."+rName, "id"),
					resource.TestCheckResourceAttrSet("cachefly_user."+rName, "status"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "cachefly_user." + rName,
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore password since it's sensitive and not returned by read
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

// Helper function to check if user exists
func testAccCheckUserExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User ID is set")
		}

		sdkClient := provider.GetSDKClient()
		if sdkClient == nil {
			return fmt.Errorf("Failed to create CacheFly client")
		}

		// Check if the user exists via API call
		_, err := sdkClient.Users.GetByID(context.Background(), rs.Primary.ID, "")
		if err != nil {
			return fmt.Errorf("User %s not found: %s", rs.Primary.ID, err.Error())
		}

		return nil
	}
}

// Helper function to check if user is destroyed
func checkUserDestroy(s *terraform.State) error {
	sdkClient := provider.GetSDKClient()
	if sdkClient == nil {
		return fmt.Errorf("Failed to create CacheFly client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cachefly_user" {
			continue
		}

		// Try to find the user
		user, err := sdkClient.Users.GetByID(context.Background(), rs.Primary.ID, "")

		if err != nil {
			return fmt.Errorf("API error when checking if user %s exists: %s", rs.Primary.ID, err.Error())
		}

		if user.Status == "DELETED" {
			return nil
		} else {
			return fmt.Errorf("User %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

// Test configuration for basic user
func testAccUserResourceConfig(name, email string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_user" %[1]q {
  username               = %[1]q
  email                  = %[2]q
  full_name              = "%[1]s User"
  phone                  = "+1234567890"
  password               = "TempPassword123!"
  password_change_required = false
}
`, name, email)
}

// Test configuration for updated user
func testAccUserResourceConfigUpdated(name, email string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_user" %[1]q {
  username               = %[1]q
  email                  = %[2]q
  full_name              = "%[1]s Updated User"
  phone                  = "+0987654321"
  password               = "UpdatedPassword123!"
  password_change_required = true
}
`, name, email)
}

// Test configuration for user with services and permissions
func testAccUserResourceConfigWithServicesAndPermissions(name, email string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" "%[1]s-svc" {
  name        = "%[1]s-svc"
  unique_name = "%[1]s-svc-unique"
  description = "%[1]s svc for user testing"
}

resource "cachefly_user" %[1]q {
  username               = %[1]q
  email                  = %[2]q
  full_name              = "%[1]s Service User"
  password               = "ServicePassword123!"
  password_change_required = false
  services               = [cachefly_service.%[1]s-svc.id]
  permissions            = ["P_USER_PROFILE_VIEW", "P_SERVICE_ALL"]
}
`, name, email)
}
