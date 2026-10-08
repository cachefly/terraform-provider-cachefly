# Account-level store
data "cachefly_edge_control_kv" "account" {}

# Store of one service
data "cachefly_edge_control_kv" "service" {
  service_id = "681b3dc52715310035cb75d4"
}

# What the service's edge scripts see: account keys overridden by service keys
data "cachefly_edge_control_kv" "merged" {
  service_id = "681b3dc52715310035cb75d4"
  merged     = true
}

output "merged_kv" {
  value = data.cachefly_edge_control_kv.merged.data
}
