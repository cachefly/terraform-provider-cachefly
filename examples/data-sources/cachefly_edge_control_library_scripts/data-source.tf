# All request scripts curated by CacheFly
data "cachefly_edge_control_library_scripts" "system_request" {
  type = "SYSTEM"
  kind = "REQUEST"
}

output "system_request_scripts" {
  value = { for s in data.cachefly_edge_control_library_scripts.system_request.scripts : s.name => s.id }
}
