package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider"
)

func TestAccServiceDataSource_ByIDAndUniqueName(t *testing.T) {
	rName := "test-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// By ID
					resource.TestCheckResourceAttrPair("data.cachefly_service.by_id", "id", "cachefly_service."+rName, "id"),
					resource.TestCheckResourceAttrPair("data.cachefly_service.by_id", "unique_name", "cachefly_service."+rName, "unique_name"),
					resource.TestCheckResourceAttrSet("data.cachefly_service.by_id", "name"),
					resource.TestCheckResourceAttrSet("data.cachefly_service.by_id", "status"),
					resource.TestCheckResourceAttrSet("data.cachefly_service.by_id", "created_at"),
					resource.TestCheckResourceAttrSet("data.cachefly_service.by_id", "updated_at"),

					// By unique_name
					resource.TestCheckResourceAttrPair("data.cachefly_service.by_unique", "id", "cachefly_service."+rName, "id"),
					resource.TestCheckResourceAttrPair("data.cachefly_service.by_unique", "unique_name", "cachefly_service."+rName, "unique_name"),
					resource.TestCheckResourceAttrSet("data.cachefly_service.by_unique", "name"),
					resource.TestCheckResourceAttrSet("data.cachefly_service.by_unique", "status"),
				),
			},
		},
	})
}

func testAccServiceDataSourceConfig(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" %[1]q {
  name        = %[1]q
  unique_name = "%[1]s-unique"
  description = "%[1]s description"
}

data "cachefly_service" "by_id" {
  id = cachefly_service.%[1]s.id
}

data "cachefly_service" "by_unique" {
  unique_name = cachefly_service.%[1]s.unique_name
}
`, name)
}
