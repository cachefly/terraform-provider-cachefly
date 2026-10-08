package resources_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider"
	"github.com/cachefly/terraform-provider-cachefly/internal/provider/resources"
)

func TestEdgeControlLibraryScriptResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := resources.NewEdgeControlLibraryScriptResource()
	resp := &fwresource.SchemaResponse{}

	r.Schema(ctx, fwresource.SchemaRequest{}, resp)

	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	attrs := resp.Schema.Attributes
	for _, name := range []string{"name", "kind", "code"} {
		assert.True(t, attrs[name].IsRequired(), "%s should be required", name)
	}
	for _, name := range []string{"id", "type", "created_at", "updated_at"} {
		assert.True(t, attrs[name].IsComputed(), "%s should be computed", name)
	}
}

func TestEdgeControlLibraryScriptResourceMetadata(t *testing.T) {
	r := resources.NewEdgeControlLibraryScriptResource()
	resp := &fwresource.MetadataResponse{}

	r.Metadata(context.Background(), fwresource.MetadataRequest{ProviderTypeName: "cachefly"}, resp)

	assert.Equal(t, "cachefly_edge_control_library_script", resp.TypeName)
}

func TestAccEdgeControlLibraryScriptResource(t *testing.T) {
	rName := "test-eclib-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "cachefly_edge_control_library_script.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccEdgeControlPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkEdgeControlLibraryScriptDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccEdgeControlLibraryScriptConfig(rName, "REQUEST", testAccEdgeControlScriptV1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "kind", "REQUEST"),
					resource.TestCheckResourceAttr(resourceName, "code", testAccEdgeControlScriptV1),
					resource.TestCheckResourceAttr(resourceName, "type", "USER"),
				),
			},
			// Update testing
			{
				Config: testAccEdgeControlLibraryScriptConfig(rName+"-renamed", "RESPONSE", testAccEdgeControlScriptV2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName+"-renamed"),
					resource.TestCheckResourceAttr(resourceName, "kind", "RESPONSE"),
					resource.TestCheckResourceAttr(resourceName, "code", testAccEdgeControlScriptV2),
				),
			},
			// ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Helper function to check that library scripts are deleted
func checkEdgeControlLibraryScriptDestroy(s *terraform.State) error {
	sdkClient := provider.GetSDKClient()
	if sdkClient == nil {
		return fmt.Errorf("Failed to create CacheFly client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cachefly_edge_control_library_script" {
			continue
		}

		_, err := sdkClient.EdgeControlLibrary.GetByID(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Edge control library script %s still exists", rs.Primary.ID)
		}
		if !strings.Contains(err.Error(), "API error 404") {
			return fmt.Errorf("API error when checking edge control library script %s: %s", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccEdgeControlLibraryScriptConfig(name string, kind string, code string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_edge_control_library_script" "test" {
  name = %[1]q
  kind = %[2]q
  code = %[3]q
}
`, name, kind, code)
}
