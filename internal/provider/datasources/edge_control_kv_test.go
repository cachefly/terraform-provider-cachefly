package datasources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider"
)

func TestAccEdgeControlKVDataSource(t *testing.T) {
	rName := "test-eckvds-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccEdgeControlPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEdgeControlKVDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.cachefly_edge_control_kv.service", "id", "cachefly_service.test", "id"),
					resource.TestCheckResourceAttr("data.cachefly_edge_control_kv.service", "data.%", "2"),
					resource.TestCheckResourceAttr("data.cachefly_edge_control_kv.service", "data.region", "eu"),
					resource.TestCheckResourceAttr("data.cachefly_edge_control_kv.service", "data.max_age", "300"),
					resource.TestCheckResourceAttrSet("data.cachefly_edge_control_kv.service", "updated_at"),
					resource.TestCheckResourceAttr("data.cachefly_edge_control_kv.merged", "data.region", "eu"),
					resource.TestCheckNoResourceAttr("data.cachefly_edge_control_kv.merged", "updated_at"),
				),
			},
		},
	})
}

func TestAccEdgeControlKVDataSourceMergedRequiresServiceID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "cachefly" {}

data "cachefly_edge_control_kv" "merged" {
  merged = true
}
`,
				ExpectError: regexp.MustCompile("Missing service_id"),
			},
		},
	})
}

func testAccEdgeControlKVDataSourceConfig(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" "test" {
  name        = %[1]q
  unique_name = "%[1]s-unique"
}

resource "cachefly_edge_control_kv" "test" {
  service_id = cachefly_service.test.id
  data = {
    region  = "eu"
    max_age = 300
  }
}

data "cachefly_edge_control_kv" "service" {
  service_id = cachefly_edge_control_kv.test.service_id
}

data "cachefly_edge_control_kv" "merged" {
  service_id = cachefly_edge_control_kv.test.service_id
  merged     = true
}
`, name)
}
