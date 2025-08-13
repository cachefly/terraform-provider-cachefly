---
page_title: "Quickstart Setup Guide"
---

# Quickstart Setup Guide

### Prerequisites

- CacheFly account and API token
- Terraform >= 1.0 installed
- A domain you control for the examples (update `var.domain`)

### Authentication

- Recommended: set environment variables so the provider can auto-configure.

```bash
export CACHEFLY_API_TOKEN="your-api-token"
# Optional override (defaults to https://api.cachefly.com/api/2.5)
export CACHEFLY_BASE_URL="https://api.cachefly.com/api/2.5"
```

```powershell
$env:CACHEFLY_API_TOKEN = "your-api-token"
$env:CACHEFLY_BASE_URL  = "https://api.cachefly.com/api/2.5"
```

Alternatively, set `api_token` in the `provider "cachefly" {}` block.

### How to run

1. Copy the example below into an empty folder as `main.tf`.
2. Initialize and apply:

```bash
terraform init
terraform apply
```

4. When done, clean up:

```bash
terraform destroy
```


```terraform
# ===================================================================
# CacheFly CDN - example demonstrating  Quickstart CDN service
# 
# ===================================================================

terraform {
  required_version = ">= 1.0"
  required_providers {
    cachefly = {
      source = "cachefly/cachefly" 
      version = "0.0.0"
    }
  }
}


provider "cachefly" {}

# Create your first CDN service
resource "cachefly_service" "my_cdn" {
  name        = "my-first-cdn"
  unique_name = "my-first-cdn"
  description = "My first CacheFly CDN service"
  auto_ssl    = true

  options = {
    reverseProxy = {
      enabled           = true
      originScheme      = "HTTP"
      cacheByQueryParam = true
      useRobotsTxt      = true
      ttl               = 3600
      hostname          = "example.com"
    }
  }
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
```