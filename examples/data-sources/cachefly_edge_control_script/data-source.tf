# Active request script of a service
data "cachefly_edge_control_script" "active" {
  service_id = "681b3dc52715310035cb75d4"
  kind       = "REQUEST"
}

# A specific published version
data "cachefly_edge_control_script" "v1" {
  service_id = "681b3dc52715310035cb75d4"
  kind       = "REQUEST"
  version    = 1
}

output "active_version" {
  value = data.cachefly_edge_control_script.active.version
}
