# CacheFly Edge Control Script Example
# Edge control scripts run JavaScript on the CDN for a service. Every change to
# `script` publishes a new, immutable version and activates it.

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

# Request script, loaded from a file next to the configuration
resource "cachefly_edge_control_script" "request" {
  service_id = cachefly_service.example.id
  kind       = "REQUEST"
  script     = file("${path.module}/edge/request.js")
}

# Response script, published but not active yet
resource "cachefly_edge_control_script" "response" {
  service_id = cachefly_service.example.id
  kind       = "RESPONSE"
  activated  = false
  script     = <<-EOT
    function handler(event) {
      return event;
    }
  EOT
}
