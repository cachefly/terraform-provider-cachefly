terraform {
  required_providers {
    cachefly = {
      source  = "cachefly.com/avvvet/cachefly"
      version = "0.1.0"
    }
  }
}

provider "cachefly" {
  api_token = "API-TOKEN" # Replace with a valid CacheFly API token
}

resource "cachefly_service" "east_server" {
  name        = "my-test-service-20250509"
  unique_name = "my-test-service-20250509"
  description = "This is a test service created by Terraform"
}

output "created_service" {
  value = cachefly_service.east_server
}