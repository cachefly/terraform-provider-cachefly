package resources_test

import (
	"context"
	"fmt"
	"os"
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

func TestEdgeControlKVResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := resources.NewEdgeControlKVResource()
	resp := &fwresource.SchemaResponse{}

	r.Schema(ctx, fwresource.SchemaRequest{}, resp)

	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	attrs := resp.Schema.Attributes
	assert.True(t, attrs["data"].IsRequired(), "data should be required")
	assert.True(t, attrs["service_id"].IsOptional(), "service_id should be optional")
	for _, name := range []string{"id", "created_at", "updated_at"} {
		assert.True(t, attrs[name].IsComputed(), "%s should be computed", name)
	}
}

func TestEdgeControlKVResourceMetadata(t *testing.T) {
	r := resources.NewEdgeControlKVResource()
	resp := &fwresource.MetadataResponse{}

	r.Metadata(context.Background(), fwresource.MetadataRequest{ProviderTypeName: "cachefly"}, resp)

	assert.Equal(t, "cachefly_edge_control_kv", resp.TypeName)
}

func TestAccEdgeControlKVResourceService(t *testing.T) {
	rName := "test-eckv-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "cachefly_edge_control_kv.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccEdgeControlPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkEdgeControlKVDestroy,
		Steps: []resource.TestStep{
			// Create with mixed value types
			{
				Config: testAccEdgeControlServiceKVConfig(rName, `{
    region      = "eu"
    max_age     = 300
    maintenance = false
  }`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", "cachefly_service.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "data.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "data.region", "eu"),
					resource.TestCheckResourceAttr(resourceName, "data.max_age", "300"),
					resource.TestCheckResourceAttr(resourceName, "data.maintenance", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			// The whole store is replaced: removed keys disappear
			{
				Config: testAccEdgeControlServiceKVConfig(rName, `{
    region = "us"
  }`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "data.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "data.region", "us"),
				),
			},
			// ImportState testing
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
		},
	})
}

// The account-level store is shared by every service of the account, so this
// test only runs when explicitly enabled.
func TestAccEdgeControlKVResourceAccount(t *testing.T) {
	if os.Getenv("CACHEFLY_ACC_ACCOUNT_KV") == "" {
		t.Skip("Set CACHEFLY_ACC_ACCOUNT_KV=1 to run: the test replaces and then clears the account-level edge control KV store")
	}

	resourceName := "cachefly_edge_control_kv.account"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccEdgeControlPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		CheckDestroy:             checkEdgeControlKVDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
provider "cachefly" {}

resource "cachefly_edge_control_kv" "account" {
  data = {
    feature_enabled = true
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "account"),
					resource.TestCheckNoResourceAttr(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, "data.feature_enabled", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateId:           "account",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
		},
	})
}

// Helper function to check that destroyed stores are empty
func checkEdgeControlKVDestroy(s *terraform.State) error {
	sdkClient := provider.GetSDKClient()
	if sdkClient == nil {
		return fmt.Errorf("Failed to create CacheFly client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cachefly_edge_control_kv" {
			continue
		}

		serviceID := rs.Primary.Attributes["service_id"]
		var data map[string]interface{}
		if serviceID == "" {
			kv, err := sdkClient.EdgeControlKV.GetAccount(context.Background())
			if err != nil {
				return fmt.Errorf("API error when checking the account-level edge control KV store: %s", err)
			}
			data = kv.Data
		} else {
			kv, err := sdkClient.EdgeControlKV.GetService(context.Background(), serviceID)
			if err != nil {
				if strings.Contains(err.Error(), "API error 404") {
					continue
				}
				return fmt.Errorf("API error when checking the edge control KV store of service %s: %s", serviceID, err)
			}
			data = kv.Data
		}

		if len(data) > 0 {
			return fmt.Errorf("Edge control KV store %s still has %d keys", rs.Primary.ID, len(data))
		}
	}

	return nil
}

func testAccEdgeControlServiceKVConfig(name string, data string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" "test" {
  name        = %[1]q
  unique_name = "%[1]s-unique"
}

resource "cachefly_edge_control_kv" "test" {
  service_id = cachefly_service.test.id
  data = %[2]s
}
`, name, data)
}
