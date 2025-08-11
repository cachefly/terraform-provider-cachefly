package resources_test

import (
	"context"
	"fmt"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"
	"github.com/cachefly/terraform-provider-cachefly/internal/provider"
	"github.com/cachefly/terraform-provider-cachefly/internal/provider/resources"
)

// Schema validation
func TestScriptConfigResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := resources.NewScriptConfigResource()
	req := fwresource.SchemaRequest{}
	resp := &fwresource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	// Verify no errors
	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	// Verify required attributes exist
	attrs := resp.Schema.Attributes
	assert.Contains(t, attrs, "name")
	assert.Contains(t, attrs, "services")
	assert.Contains(t, attrs, "script_config_definition")

	// optional attributes exist
	assert.Contains(t, attrs, "mime_type")
	assert.Contains(t, attrs, "value")
	assert.Contains(t, attrs, "activated")

	// computed attributes exist
	assert.Contains(t, attrs, "id")
	assert.Contains(t, attrs, "purpose")
	assert.Contains(t, attrs, "created_at")
	assert.Contains(t, attrs, "updated_at")

	// required attributes are marked as required
	assert.True(t, attrs["name"].IsRequired())
	assert.True(t, attrs["script_config_definition"].IsRequired())

	// optional attributes are marked as optional
	assert.True(t, attrs["mime_type"].IsOptional())
	assert.True(t, attrs["value"].IsOptional())
	assert.True(t, attrs["activated"].IsOptional())

	// computed attributes are marked as computed
	assert.True(t, attrs["id"].IsComputed())
	assert.True(t, attrs["purpose"].IsComputed())
	assert.True(t, attrs["created_at"].IsComputed())
	assert.True(t, attrs["updated_at"].IsComputed())

	// attributes with defaults are also computed
	assert.True(t, attrs["mime_type"].IsComputed())
	assert.True(t, attrs["activated"].IsComputed())
}

// Resource metadata
func TestScriptConfigResourceMetadata(t *testing.T) {
	ctx := context.Background()
	r := resources.NewScriptConfigResource()
	req := fwresource.MetadataRequest{
		ProviderTypeName: "cachefly",
	}
	resp := &fwresource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "cachefly_script_config", resp.TypeName)
}

// Configure error handling
func TestScriptConfigResourceConfigure(t *testing.T) {
	ctx := context.Background()
	r := resources.NewScriptConfigResource().(*resources.ScriptConfigResource)

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

// Acceptance test for script_config resource
func TestAccScriptConfigResource(t *testing.T) {
	sdkClient := provider.GetSDKClient()

	// Fetch available script config definitions and take the first one
	defs, err := sdkClient.ScriptDefinitions.List(context.Background(), v2_5.ListScriptDefinitionsOptions{})
	if err != nil {
		t.Fatalf("Failed to get script config definitions: %v", err)
	}
	if len(defs.Definitions) == 0 {
		t.Fatalf("No script config definitions available for the account")
	}
	definitionID := defs.Definitions[0].ID

	rName := "test-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "cachefly_script_config." + rName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkScriptConfigDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccScriptConfigResourceConfig(rName, definitionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScriptConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "script_config_definition", definitionID),
					resource.TestCheckResourceAttr(resourceName, "value", "test"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at", "value"},
			},
		},
	})
}

// Helper to check if script config exists
func testAccCheckScriptConfigExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ScriptConfig ID is set")
		}

		sdkClient := provider.GetSDKClient()
		if sdkClient == nil {
			return fmt.Errorf("Failed to create CacheFly client")
		}

		// Check if the script config exists via API call
		_, err := sdkClient.ScriptConfigs.GetByID(context.Background(), rs.Primary.ID, "")
		if err != nil {
			return fmt.Errorf("ScriptConfig %s not found: %s", rs.Primary.ID, err.Error())
		}

		return nil
	}
}

// Helper to verify script config is destroyed (deactivated)
func checkScriptConfigDestroy(s *terraform.State) error {
	sdkClient := provider.GetSDKClient()
	if sdkClient == nil {
		return fmt.Errorf("Failed to create CacheFly client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cachefly_script_config" {
			continue
		}

		cfg, err := sdkClient.ScriptConfigs.GetByID(context.Background(), rs.Primary.ID, "")
		if err != nil {
			return fmt.Errorf("API error when checking if script config %s exists: %s", rs.Primary.ID, err.Error())
		}

		if cfg.Status == "DEACTIVATED" {
			return nil
		}
		return fmt.Errorf("Script config %s still exists", rs.Primary.ID)
	}

	return nil
}

// Test configuration for basic script config
func testAccScriptConfigResourceConfig(name string, definitionID string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_script_config" %q {
  name                     = %q
  script_config_definition = %q
  value                    = <<-JSON5
  {
    something: 2324,
  }
  JSON5
}
`, name, name, definitionID)
}
