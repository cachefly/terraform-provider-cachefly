package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider"
)

func TestAccEdgeControlLibraryScriptDataSources(t *testing.T) {
	rName := "test-eclibds-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccEdgeControlPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEdgeControlLibraryScriptDataSourcesConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.cachefly_edge_control_library_script.test", "name", rName),
					resource.TestCheckResourceAttr("data.cachefly_edge_control_library_script.test", "kind", "REQUEST"),
					resource.TestCheckResourceAttr("data.cachefly_edge_control_library_script.test", "type", "USER"),
					resource.TestCheckResourceAttrPair("data.cachefly_edge_control_library_script.test", "code", "cachefly_edge_control_library_script.test", "code"),
					resource.TestCheckResourceAttr("data.cachefly_edge_control_library_scripts.user", "scripts.#", "1"),
					resource.TestCheckResourceAttrPair("data.cachefly_edge_control_library_scripts.user", "scripts.0.id", "cachefly_edge_control_library_script.test", "id"),
					resource.TestCheckResourceAttr("data.cachefly_edge_control_library_scripts.user", "scripts.0.type", "USER"),
				),
			},
		},
	})
}

func testAccEdgeControlLibraryScriptDataSourcesConfig(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_edge_control_library_script" "test" {
  name = %[1]q
  kind = "REQUEST"
  code = "function handler(event) {\n  return event;\n}"
}

data "cachefly_edge_control_library_script" "test" {
  id = cachefly_edge_control_library_script.test.id
}

data "cachefly_edge_control_library_scripts" "user" {
  type   = "USER"
  kind   = cachefly_edge_control_library_script.test.kind
  search = cachefly_edge_control_library_script.test.name
}
`, name)
}
