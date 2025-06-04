# CacheFly Provider - Complete Setup Example
# This example shows services with custom domains - a real-world setup

terraform {
  required_version = ">= 1.0"
  required_providers {
    cachefly = {
      source = "cachefly.com/avvvet/cachefly" # todo: cachefly/cachefly
    }
  }
}

provider "cachefly" {
  api_token = ""
}

# Variables for configuration
variable "project_name" {
  description = "Name of your project"
  type        = string
  default     = "my-app"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  default     = "prod"
}

variable "base_domain" {
  description = "Your base domain (e.g., example.com)"
  type        = string
  default     = "thobingo.online"
}

# Local values for computed configurations
locals {
  name_prefix = "${var.project_name}-${var.environment}"
}

# ===================================================================
# SERVICES - CDN Services
# ===================================================================

# Main web application service
resource "cachefly_service" "web_app" {
  name               = "${local.name_prefix}-web"
  unique_name        = "${local.name_prefix}-web-05"
  description        = "Web application CDN for ${var.project_name}"
  auto_ssl           = true
  configuration_mode = "API_RULES_AND_OPTIONS"
}

# API service 
resource "cachefly_service" "api" {
  name               = "${local.name_prefix}-api"
  unique_name        = "${local.name_prefix}-api-05"
  description        = "API CDN for ${var.project_name}"
  auto_ssl           = true
  configuration_mode = "API_RULES_AND_OPTIONS"
}

# Static assets service
resource "cachefly_service" "assets" {
  name               = "${local.name_prefix}-assets"
  unique_name        = "${local.name_prefix}-assets-05"
  description        = "Static assets CDN for ${var.project_name}"
  auto_ssl           = true
  configuration_mode = "API_RULES_AND_OPTIONS"
}

# ===================================================================
# SERVICE DOMAINS - Custom Domains
# ===================================================================

# Main website domain
resource "cachefly_service_domain" "web_main" {
  service_id       = cachefly_service.web_app.id
  name             = var.base_domain
  description      = "Main website domain"
  validation_mode  = "DNS"
}

# www subdomain for web
resource "cachefly_service_domain" "web_www" {
  service_id       = cachefly_service.web_app.id
  name             = "www.${var.base_domain}"
  description      = "WWW subdomain for website"
  validation_mode  = "DNS"
}

# API subdomain
resource "cachefly_service_domain" "api_subdomain" {
  service_id       = cachefly_service.api.id
  name             = "api.${var.base_domain}"
  description      = "API subdomain"
  validation_mode  = "DNS"
}

# CDN subdomain for assets
resource "cachefly_service_domain" "assets_cdn" {
  service_id       = cachefly_service.assets.id
  name             = "cdn.${var.base_domain}"
  description      = "CDN subdomain for static assets"
  validation_mode  = "DNS"
}

# ===================================================================
# DATA SOURCES - Lookup and Verification
# ===================================================================

# Look up services to verify they were created
data "cachefly_service" "web_app_verify" {
  id = cachefly_service.web_app.id
}

# Look up domains for the web service
data "cachefly_service_domains" "web_domains" {
  service_id = cachefly_service.web_app.id
}

# Look up specific domain
data "cachefly_service_domain" "main_domain_verify" {
  service_id = cachefly_service.web_app.id
  id         = cachefly_service_domain.web_main.id
}

# ===================================================================
# OUTPUTS - Useful Information
# ===================================================================

output "services_summary" {
  description = "Summary of all created services"
  value = {
    web_app = {
      id          = cachefly_service.web_app.id
      name        = cachefly_service.web_app.name
      unique_name = cachefly_service.web_app.unique_name
      status      = cachefly_service.web_app.status
      ssl_enabled = cachefly_service.web_app.auto_ssl
    }
    api = {
      id          = cachefly_service.api.id
      name        = cachefly_service.api.name
      unique_name = cachefly_service.api.unique_name
      status      = cachefly_service.api.status
      ssl_enabled = cachefly_service.api.auto_ssl
    }
    assets = {
      id          = cachefly_service.assets.id
      name        = cachefly_service.assets.name
      unique_name = cachefly_service.assets.unique_name
      status      = cachefly_service.assets.status
      ssl_enabled = cachefly_service.assets.auto_ssl
    }
  }
}

output "domains_summary" {
  description = "Summary of all attached domains"
  value = {
    web_domains = {
      main = {
        id                = cachefly_service_domain.web_main.id
        name              = cachefly_service_domain.web_main.name
        validation_status = cachefly_service_domain.web_main.validation_status
      }
      www = {
        id                = cachefly_service_domain.web_www.id
        name              = cachefly_service_domain.web_www.name
        validation_status = cachefly_service_domain.web_www.validation_status
      }
    }
    api_domain = {
      id                = cachefly_service_domain.api_subdomain.id
      name              = cachefly_service_domain.api_subdomain.name
      validation_status = cachefly_service_domain.api_subdomain.validation_status
    }
    assets_domain = {
      id                = cachefly_service_domain.assets_cdn.id
      name              = cachefly_service_domain.assets_cdn.name
      validation_status = cachefly_service_domain.assets_cdn.validation_status
    }
  }
}

output "cdn_endpoints" {
  description = "Your CDN endpoints - use these in your applications"
  value = {
    website = {
      main_domain    = var.base_domain
      www_domain     = "www.${var.base_domain}"
      cachefly_url   = "https://${cachefly_service.web_app.unique_name}.cachefly.net"
    }
    api = {
      api_domain     = "api.${var.base_domain}"
      cachefly_url   = "https://${cachefly_service.api.unique_name}.cachefly.net"
    }
    assets = {
      cdn_domain     = "cdn.${var.base_domain}"
      cachefly_url   = "https://${cachefly_service.assets.unique_name}.cachefly.net"
    }
  }
}

output "verification" {
  description = "Verification that everything is working correctly"
  value = {
    all_services_active = alltrue([
      cachefly_service.web_app.status == "ACTIVE",
      cachefly_service.api.status == "ACTIVE",
      cachefly_service.assets.status == "ACTIVE"
    ])
    web_domains_count = length(data.cachefly_service_domains.web_domains.domains)
    data_sources_work = data.cachefly_service_domain.main_domain_verify.name == cachefly_service_domain.web_main.name
    
    setup_complete = alltrue([
      cachefly_service.web_app.status == "ACTIVE",
      cachefly_service.api.status == "ACTIVE", 
      cachefly_service.assets.status == "ACTIVE",
      length(data.cachefly_service_domains.web_domains.domains) >= 2
    ])
  }
}

