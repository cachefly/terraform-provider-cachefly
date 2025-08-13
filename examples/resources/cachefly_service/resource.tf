# ===================================================================
# CacheFly CDN - Service Setup Examples
# ===================================================================

terraform {
  required_version = ">= 1.0"
  required_providers {
    cachefly = {
      source = "cachefly/cachefly"
    }
  }
}

provider "cachefly" {
  # Uses CACHEFLY_API_TOKEN if set; otherwise, set api_token here
  # api_token = ""
}

# -------------------------------------------------------------------
# Data Sources: Delivery Regions (optional helper)
# -------------------------------------------------------------------
data "cachefly_delivery_regions" "available" {}

# -------------------------------------------------------------------
# Example 1: Basic Service
# -------------------------------------------------------------------
resource "cachefly_service" "basic" {
  name        = "example-dev-svc"
  unique_name = "example-dev-svc-basic"
  description = "Basic CDN service example"
  auto_ssl    = true
}

# -------------------------------------------------------------------
# Example 2: Service with Options and Delivery Region
# -------------------------------------------------------------------
resource "cachefly_service" "with_options" {
  name        = "example-dev-svc-opts"
  unique_name = "example-dev-svc-opts"
  description = "Service with advanced options enabled"
  auto_ssl    = true

  # Optionally pin the service to a delivery region (pick first available)
  delivery_region = data.cachefly_delivery_regions.available.regions[0].id

  # Optionally set a TLS profile if you have a specific profile ID
  # tls_profile = "<tls-profile-id>"

  # Example options (see Service Options Reference in docs)
  options = {
    autoRedirect = true

    reverseProxy = {
      enabled           = true
      originScheme      = "HTTP"
      cacheByQueryParam = true
      useRobotsTxt      = true
      ttl               = 123
      hostname          = "example.com"
    }

    protectServeKeyEnabled = true

    purgemode = {
      enabled = true
      value = {
        exact     = true
        directory = true
        extension = true
      }
    }
  }
}

# -------------------------------------------------------------------
# Outputs
# -------------------------------------------------------------------
output "services" {
  value = {
    basic = {
      id          = cachefly_service.basic.id
      name        = cachefly_service.basic.name
      unique_name = cachefly_service.basic.unique_name
      status      = cachefly_service.basic.status
      created     = cachefly_service.basic.created_at
      updated     = cachefly_service.basic.updated_at
    }
    with_options = {
      id                  = cachefly_service.with_options.id
      name                = cachefly_service.with_options.name
      unique_name         = cachefly_service.with_options.unique_name
      status              = cachefly_service.with_options.status
      configuration_mode  = cachefly_service.with_options.configuration_mode
      delivery_region     = cachefly_service.with_options.delivery_region
      created             = cachefly_service.with_options.created_at
      updated             = cachefly_service.with_options.updated_at
    }
  }
}


