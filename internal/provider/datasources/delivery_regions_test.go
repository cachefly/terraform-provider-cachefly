package datasources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider"
)

func TestAccDeliveryRegionsDataSource_List(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryRegionsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.cachefly_delivery_regions.all", "regions.#"),
				),
			},
		},
	})
}

func testAccDeliveryRegionsDataSourceConfig() string {
	return `
provider "cachefly" {}

data "cachefly_delivery_regions" "all" {}
`
}
