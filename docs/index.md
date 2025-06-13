---
page_title: "Provider: CacheFly"
description: |-
  The CacheFly provider is used to interact with CacheFly CDN resources.
---

# CacheFly Provider

The CacheFly provider is used to interact with CacheFly CDN resources. It provides resources to manage CDN services, origins, domains, SSL certificates, users, and advanced configurations.

## Example Usage

```terraform
terraform {
  required_providers {
    cachefly = {
      source  = "cachefly/cachefly"
      version = "0.1.0"
    }
  }
}

provider "cachefly" {
  api_token = var.cachefly_api_token
}

# Create an origin server
resource "cachefly_origin" "web_server" {
  type                       = "WEB"
  name                       = "main-web-server"
  host                       = "web.example.com"
  scheme                     = "HTTPS"
  cache_by_query_param       = false
  gzip                       = true
  ttl                        = 86400
  missed_ttl                 = 300
  connection_timeout         = 15
  time_to_first_byte_timeout = 15
}

# Create a CDN service
resource "cachefly_service" "website" {
  name               = "my-website"
  unique_name        = "my-website-unique"
  description        = "Main website CDN service"
  auto_ssl           = true
  configuration_mode = "API_RULES_AND_OPTIONS"
}

# Add a custom domain
resource "cachefly_service_domain" "main_domain" {
  service_id      = cachefly_service.website.id
  name            = "example.com"
  description     = "Main website domain"
  validation_mode = "DNS"
}
```

## Authentication

The CacheFly provider requires an API token for authentication. You can obtain an API token from your CacheFly account dashboard.

### Environment Variables

You can provide your credentials via the `CACHEFLY_API_TOKEN` environment variable:

```shell
export CACHEFLY_API_TOKEN="your-api-token-here"
terraform plan
```

### Static Configuration

```terraform
provider "cachefly" {
  api_token = "your-api-token-here"
}
```

**Warning:** Hard-coding credentials into any Terraform configuration is not recommended. Use environment variables or Terraform variables instead.

## Schema

### Required

- `api_token` (String, Sensitive) - The API token for CacheFly authentication. Can also be set via the `CACHEFLY_API_TOKEN` environment variable.

## Resources

The CacheFly provider supports the following resources:

- [`cachefly_service`](resources/service.md) - Manage CDN services
- [`cachefly_service_domain`](resources/service_domain.md) - Manage custom domains for services
- [`cachefly_origin`](resources/origin.md) - Manage origin servers
- [`cachefly_service_options`](resources/service_options.md) - Manage advanced service configurations
- [`cachefly_certificate`](resources/certificate.md) - Manage SSL certificates
- [`cachefly_script_config`](resources/script_config.md) - Manage script configurations and automation
- [`cachefly_user`](resources/user.md) - Manage user accounts and permissions

## Data Sources

The CacheFly provider supports the following data sources:

- [`cachefly_origins`](data-sources/origins.md) - Retrieve information about origins
- [`cachefly_service`](data-sources/service.md) - Retrieve information about a service
- [`cachefly_service_domains`](data-sources/service_domains.md) - Retrieve information about service domains
- [`cachefly_service_options`](data-sources/service_options.md) - Retrieve configuration options for a service
- [`cachefly_users`](data-sources/users.md) - Retrieve information about users

## Common Use Cases

### Basic CDN Setup

```terraform
# 1. Create origin server
resource "cachefly_origin" "app_origin" {
  type   = "WEB"
  name   = "app-server"
  host   = "app.example.com"
  scheme = "HTTPS"
  gzip   = true
  ttl    = 86400
}

# 2. Create CDN service
resource "cachefly_service" "app_service" {
  name               = "app-cdn"
  unique_name        = "app-cdn-unique"
  auto_ssl           = true
  configuration_mode = "API_RULES_AND_OPTIONS"
}

# 3. Add custom domain
resource "cachefly_service_domain" "app_domain" {
  service_id      = cachefly_service.app_service.id
  name            = "cdn.example.com"
  validation_mode = "DNS"
}
```

### Multi-Environment Setup

```terraform
locals {
  environments = ["dev", "staging", "prod"]
}

resource "cachefly_service" "app" {
  for_each = toset(local.environments)
  
  name               = "app-${each.key}"
  unique_name        = "app-${each.key}-${random_string.suffix.result}"
  description        = "CDN service for ${each.key} environment"
  auto_ssl           = true
  configuration_mode = "API_RULES_AND_OPTIONS"
}
```

### User Management

```terraform
resource "cachefly_user" "developer" {
  username  = "developer"
  email     = "dev@example.com"
  full_name = "Developer User"
  password  = var.developer_password
  
  services = [cachefly_service.website.id]
  
  permissions = [
    "P_ACCOUNT_VIEW",
    "P_SERVICE_PURGE"
  ]
}
```