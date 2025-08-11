# ===================================================================
# CacheFly CDN - example demonstrating  Quickstart CDN service
# 
# ===================================================================

terraform {
  required_version = ">= 1.0"
  required_providers {
    cachefly = {
      source  = "cachefly/cachefly"
      version = "0.0.0"
    }
  }
}


provider "cachefly" {}

# Create your first CDN service
resource "cachefly_service" "my_cdn" {
  name        = "my-first-cdn-marko-5"
  unique_name = "my-first-cdn-marko-5"
  description = "My first CacheFly CDN service"
  auto_ssl    = true
  status      = "ACTIVE"
  options = {
    "TestOption" = false
  }
}

# Output the service information
output "my_cdn_info" {
  value = {
    service_id = cachefly_service.my_cdn.id
    status     = cachefly_service.my_cdn.status
    auto_ssl   = cachefly_service.my_cdn.auto_ssl
    created_at = cachefly_service.my_cdn.created_at
  }
}
