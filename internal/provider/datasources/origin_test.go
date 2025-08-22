package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider"
)

func TestAccOriginDataSource_Basic(t *testing.T) {
	rName := "test-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOriginDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.cachefly_origin.test", "id", "cachefly_origin."+rName, "id"),
					resource.TestCheckResourceAttr("data.cachefly_origin.test", "type", "WEB"),
					resource.TestCheckResourceAttr("data.cachefly_origin.test", "hostname", "example.com"),
				),
			},
		},
	})
}

func TestAccOriginDataSource_S3(t *testing.T) {
	rName := "test-s3-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { provider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOriginDataSourceConfigS3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.cachefly_origin.test", "id", "cachefly_origin."+rName, "id"),
					resource.TestCheckResourceAttr("data.cachefly_origin.test", "type", "S3_BUCKET"),
					resource.TestCheckResourceAttr("data.cachefly_origin.test", "hostname", "my-bucket.s3.amazonaws.com"),
					resource.TestCheckResourceAttr("data.cachefly_origin.test", "region", "us-east-1"),
					resource.TestCheckResourceAttr("data.cachefly_origin.test", "signature_version", "v4"),
				),
			},
		},
	})
}

func testAccOriginDataSourceConfig(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_origin" %[1]q {
  name                   = %[1]q
  type                   = "WEB"
  hostname               = "example.com"
  scheme                 = "HTTPS"
}

data "cachefly_origin" "test" {
  id = cachefly_origin.%[1]s.id
}
`, name)
}

func testAccOriginDataSourceConfigS3(name string) string {
	return fmt.Sprintf(`
provider "cachefly" {}

resource "cachefly_origin" %[1]q {
  name              = %[1]q
  type              = "S3_BUCKET"
  hostname          = "my-bucket.s3.amazonaws.com"
  access_key        = "AKIAIOSFODNN7EXAMPLE"
  secret_key        = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  region            = "us-east-1"
  signature_version = "v4"
}

data "cachefly_origin" "test" {
  id = cachefly_origin.%[1]s.id
}
`, name)
}
