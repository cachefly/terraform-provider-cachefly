package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider"
)

func TestAccEdgeControlScriptDataSources(t *testing.T) {
	rName := "test-ecds-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccEdgeControlPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEdgeControlScriptDataSourcesConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.cachefly_edge_control_script.active", "version", "1"),
					resource.TestCheckResourceAttr("data.cachefly_edge_control_script.active", "status", "ACTIVE"),
					resource.TestCheckResourceAttrPair("data.cachefly_edge_control_script.active", "script", "cachefly_edge_control_script.test", "script"),
					resource.TestCheckResourceAttr("data.cachefly_edge_control_script.first", "version", "1"),
					resource.TestCheckResourceAttrSet("data.cachefly_edge_control_script.first", "filename"),
					resource.TestCheckResourceAttr("data.cachefly_edge_control_script_versions.all", "versions.#", "1"),
					resource.TestCheckResourceAttr("data.cachefly_edge_control_script_versions.all", "versions.0.version", "1"),
					resource.TestCheckResourceAttr("data.cachefly_edge_control_script_versions.all", "versions.0.status", "ACTIVE"),
				),
			},
		},
	})
}

func testAccEdgeControlScriptDataSourcesConfig(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" "test" {
  name        = %[1]q
  unique_name = "%[1]s-unique"
}

resource "cachefly_edge_control_script" "test" {
  service_id = cachefly_service.test.id
  kind       = "REQUEST"
  script     = "function handler(event) {\n  return event;\n}"
}

data "cachefly_edge_control_script" "active" {
  service_id = cachefly_edge_control_script.test.service_id
  kind       = cachefly_edge_control_script.test.kind
}

data "cachefly_edge_control_script" "first" {
  service_id = cachefly_edge_control_script.test.service_id
  kind       = cachefly_edge_control_script.test.kind
  version    = 1
}

data "cachefly_edge_control_script_versions" "all" {
  service_id = cachefly_edge_control_script.test.service_id
  kind       = cachefly_edge_control_script.test.kind
}
`, name)
}
