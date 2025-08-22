package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider"
)

func TestAccUsersDataSource_List(t *testing.T) {
	rName := "test-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	userEmail := rName + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUsersDataSourceConfig(rName, userEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.cachefly_users.all", "users.#"),
				),
			},
		},
	})
}

func testAccUsersDataSourceConfig(name, email string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_user" %[1]q {
  username = %[1]q
  email    = %[2]q
  full_name = "%[1]s User"
  password  = "TempPassword123!"
}

data "cachefly_users" "all" {}
`, name, email)
}
