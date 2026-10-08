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

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider"
	"github.com/cachefly/terraform-provider-cachefly/internal/provider/resources"
)

const (
	testAccEdgeControlScriptV1 = "function handler(event) {\n  return event;\n}"
	testAccEdgeControlScriptV2 = "function handler(event) {\n  // v2\n  return event;\n}"
)

func TestEdgeControlScriptResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := resources.NewEdgeControlScriptResource()
	resp := &fwresource.SchemaResponse{}

	r.Schema(ctx, fwresource.SchemaRequest{}, resp)

	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	attrs := resp.Schema.Attributes
	for _, name := range []string{"service_id", "kind", "script"} {
		assert.True(t, attrs[name].IsRequired(), "%s should be required", name)
	}
	assert.True(t, attrs["activated"].IsOptional(), "activated should be optional")
	assert.True(t, attrs["activated"].IsComputed(), "activated should be computed (it has a default)")
	for _, name := range []string{"id", "version", "status", "script_size", "filename", "last_activated_at", "created_at"} {
		assert.True(t, attrs[name].IsComputed(), "%s should be computed", name)
	}
}

func TestEdgeControlScriptResourceMetadata(t *testing.T) {
	r := resources.NewEdgeControlScriptResource()
	resp := &fwresource.MetadataResponse{}

	r.Metadata(context.Background(), fwresource.MetadataRequest{ProviderTypeName: "cachefly"}, resp)

	assert.Equal(t, "cachefly_edge_control_script", resp.TypeName)
}

func TestAccEdgeControlScriptResource(t *testing.T) {
	rName := "test-ec-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "cachefly_edge_control_script.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccEdgeControlPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkEdgeControlScriptDestroy,
		Steps: []resource.TestStep{
			// Publish and activate the first version
			{
				Config: testAccEdgeControlScriptResourceConfig(rName, testAccEdgeControlScriptV1, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "service_id", "cachefly_service.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "kind", "REQUEST"),
					resource.TestCheckResourceAttr(resourceName, "script", testAccEdgeControlScriptV1),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "activated", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "filename"),
					resource.TestCheckResourceAttrSet(resourceName, "last_activated_at"),
				),
			},
			// Changing the script publishes and activates a new version
			{
				Config: testAccEdgeControlScriptResourceConfig(rName, testAccEdgeControlScriptV2, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "script", testAccEdgeControlScriptV2),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
				),
			},
			// Deactivating keeps the version and leaves nothing active
			{
				Config: testAccEdgeControlScriptResourceConfig(rName, testAccEdgeControlScriptV2, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
					resource.TestCheckResourceAttr(resourceName, "status", "DEACTIVATED"),
					resource.TestCheckResourceAttr(resourceName, "activated", "false"),
				),
			},
			// ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccEdgeControlScriptImportID(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccEdgeControlScriptImportID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return rs.Primary.Attributes["service_id"] + ":" + rs.Primary.Attributes["kind"], nil
	}
}

// Helper function to check that no edge control script is active after destroy
func checkEdgeControlScriptDestroy(s *terraform.State) error {
	sdkClient := provider.GetSDKClient()
	if sdkClient == nil {
		return fmt.Errorf("Failed to create CacheFly client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cachefly_edge_control_script" {
			continue
		}

		serviceID := rs.Primary.Attributes["service_id"]
		kind := v2_6.EdgeControlScriptKind(rs.Primary.Attributes["kind"])
		active, err := sdkClient.EdgeControlScripts.GetActiveVersion(context.Background(), serviceID, kind)
		if err != nil {
			if strings.Contains(err.Error(), "API error 404") {
				continue
			}
			return fmt.Errorf("API error when checking the %s script of service %s: %s", kind, serviceID, err)
		}
		return fmt.Errorf("Version %d of the %s script of service %s is still active", active.Version, kind, serviceID)
	}

	return nil
}

func testAccEdgeControlScriptResourceConfig(name string, script string, activated bool) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" "test" {
  name        = %[1]q
  unique_name = "%[1]s-unique"
}

resource "cachefly_edge_control_script" "test" {
  service_id = cachefly_service.test.id
  kind       = "REQUEST"
  script     = %[2]q
  activated  = %[3]t
}
`, name, script, activated)
}
