# CacheFly Terraform Provider - Quickstart
# Get started with CacheFly in under 5 minutes

terraform {
  required_version = ">= 1.0"
  required_providers {
    cachefly = {
      source = "cachefly.com/avvvet/cachefly"
    }
  }
}

provider "cachefly" {
  api_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjQ0NDMxLCJ1c2VyIjoiNjgxYjNjZmIyNzE1MzEwMDM1Y2I3MmI2IiwidG9rZW4iOiI2ODM5YzExMjc4M2NmOTAwNDA1NzY3ZDIiLCJpYXQiOjE3NDg2MTU0NDJ9.6ZU6QW9UVqMkLTKbWr4z1o73BvaA3OqDZqtbu8k353c" 
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