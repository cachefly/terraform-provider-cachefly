# CacheFly Edge Control KV Example
# Edge control scripts can read key/value data. The account-level store applies
# to every service; a service-level store overrides account keys for one service.

terraform {
  required_providers {
    cachefly = {
      source = "cachefly/cachefly"
    }
  }
}

provider "cachefly" {
  api_token = ""
}

resource "cachefly_service" "example" {
  name        = "edge-control-example"
  unique_name = "edge-control-example"
}

# Account-level store (no service_id)
resource "cachefly_edge_control_kv" "account" {
  data = {
    maintenance = false
    max_age     = 3600
  }
}

# Service-level store; its keys override the account keys for this service
resource "cachefly_edge_control_kv" "service" {
  service_id = cachefly_service.example.id
  data = {
    region  = "eu-west"
    max_age = 300
  }
}
