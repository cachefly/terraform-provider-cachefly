data "cachefly_edge_control_script_versions" "request" {
  service_id = "681b3dc52715310035cb75d4"
  kind       = "REQUEST"
}

output "published_versions" {
  value = [for v in data.cachefly_edge_control_script_versions.request.versions : "${v.version}: ${v.status}"]
}
