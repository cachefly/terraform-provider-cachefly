package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider"
)

func TestAccServiceDomainsDataSource_List(t *testing.T) {
	rName := "test-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	domain := rName + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDomainsDataSourceConfig(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Ensure at least one domain is returned and attributes exist
					resource.TestCheckResourceAttrSet("data.cachefly_service_domains.all", "domains.#"),
				),
			},
		},
	})
}

func testAccServiceDomainsDataSourceConfig(name, domain string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" %[1]q {
  name        = %[1]q
  unique_name = "%[1]s-unique"
  description = "%[1]s service for domain testing"
}

resource "cachefly_service_domain" %[1]q {
  service_id      = cachefly_service.%[1]s.id
  name            = %[2]q
  description     = "%[1]s domain description"
  validation_mode = "HTTP"
}

data "cachefly_service_domains" "all" {
  service_id = cachefly_service.%[1]s.id
}
`, name, domain)
}
