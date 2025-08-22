package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider"
)

func TestAccLogTargetsDataSource_List(t *testing.T) {
	rName := "test-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLogTargetsDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.cachefly_log_targets.all", "log_targets.#"),
				),
			},
		},
	})
}

func testAccLogTargetsDataSourceConfig(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_service" %[1]q {
  name        = %[1]q
  unique_name = "%[1]s-unique"
  description = "%[1]s service for logs testing"
}

resource "cachefly_log_target" %[1]q {
  name                 = %[1]q
  type                 = "S3_BUCKET"
  bucket               = "my-log-bucket"
  region               = "us-east-1"
  access_key           = "AKIAIOSFODNN7EXAMPLE"
  secret_key           = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  signature_version    = "v4"
  access_logs_services = [cachefly_service.%[1]s.id]
  origin_logs_services = [cachefly_service.%[1]s.id]
}

data "cachefly_log_targets" "all" {}
`, name)
}
