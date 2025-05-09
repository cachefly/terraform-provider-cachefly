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

#data "cachefly_services" "example" {}

data "cachefly_services" "list_services" {
  response_type    = "shallow"
  include_features = false
  status           = "ACTIVE"
  offset           = 0
  limit            = 10
}

output "services" {
  value = data.cachefly_services.list_services.services
}