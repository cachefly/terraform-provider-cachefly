# CacheFly Edge Control Library Script Example
# Library scripts are reusable scripts stored in the account. They do not run by
# themselves; deploy their code with cachefly_edge_control_script.

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

resource "cachefly_edge_control_library_script" "pass_through" {
  name = "pass-through"
  kind = "REQUEST"
  code = <<-EOT
    function handler(event) {
      return event;
    }
  EOT
}

resource "cachefly_service" "example" {
  name        = "edge-control-example"
  unique_name = "edge-control-example"
}

# Deploy the library script to a service
resource "cachefly_edge_control_script" "request" {
  service_id = cachefly_service.example.id
  kind       = cachefly_edge_control_library_script.pass_through.kind
  script     = cachefly_edge_control_library_script.pass_through.code
}
