data "cachefly_edge_control_library_script" "example" {
  id = "65f1c2a9e4b0a1b2c3d4e5f6"
}

# Deploy a library script to a service
resource "cachefly_edge_control_script" "from_library" {
  service_id = "681b3dc52715310035cb75d4"
  kind       = data.cachefly_edge_control_library_script.example.kind
  script     = data.cachefly_edge_control_library_script.example.code
}
