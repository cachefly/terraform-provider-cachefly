# ===================================================================
# CacheFly CDN - example demonstrating  Quickstart CDN service
# 
# ===================================================================

terraform {
  required_version = ">= 1.0"
  required_providers {
    cachefly = {
      source = "cachefly.com/avvvet/cachefly" # todo: cachefly/cachefly
    }
  }
}



provider "cachefly" {
  api_token = "your-token" 
}

# Create your first CDN service
resource "cachefly_service" "my_cdn" {
  name        = "my-first-cdn"
  unique_name = "my-first-cdn-01"
  description = "My first CacheFly CDN service"
  auto_ssl    = true
}

# Output the service information
output "my_cdn_info" {
  value = {
    service_id  = cachefly_service.my_cdn.id
    status      = cachefly_service.my_cdn.status
    ssl_enabled = cachefly_service.my_cdn.auto_ssl
    created_at  = cachefly_service.my_cdn.created_at
  }
}