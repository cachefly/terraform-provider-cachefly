# CacheFly Provider - Complete Setup Example
# This example shows a more comprehensive service setup

terraform {
  required_version = ">= 1.0"
  required_providers {
    cachefly = {
      source = "cachefly.com/avvvet/cachefly"
    }
  }
}

provider "cachefly" {
  #"your-api-token-here"
  api_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjQ0NDMxLCJ1c2VyIjoiNjgxYjNjZmIyNzE1MzEwMDM1Y2I3MmI2IiwidG9rZW4iOiI2ODM5YzExMjc4M2NmOTAwNDA1NzY3ZDIiLCJpYXQiOjE3NDg2MTU0NDJ9.6ZU6QW9UVqMkLTKbWr4z1o73BvaA3OqDZqtbu8k353c" 
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
  unique_name        = "${local.name_prefix}-web-01"
  description        = "Web application CDN for ${var.project_name} ${var.environment}"
  auto_ssl           = true
  configuration_mode = "API_RULES_AND_OPTIONS"
}

# API service 
resource "cachefly_service" "api" {
  name               = "${local.name_prefix}-api"
  unique_name        = "${local.name_prefix}-api-01"
  description        = "API CDN for ${var.project_name} ${var.environment}"
  auto_ssl           = true
  configuration_mode = "API_RULES_AND_OPTIONS"
}

# Static assets service
resource "cachefly_service" "assets" {
  name               = "${local.name_prefix}-assets"
  unique_name        = "${local.name_prefix}-assets-01"
  description        = "Static assets CDN for ${var.project_name} ${var.environment}"
  auto_ssl           = false  # Maybe assets don't need SSL
  configuration_mode = "API_RULES_AND_OPTIONS"
}

# ===================================================================
# DATA SOURCES - Lookup and Verification
# ===================================================================

# Look up services to verify they were created
data "cachefly_service" "web_app_verify" {
  id = cachefly_service.web_app.id
}

data "cachefly_service" "api_verify" {
  unique_name = cachefly_service.api.unique_name
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

output "cdn_endpoints" {
  description = "CDN endpoints for your application"
  value = {
    web_app = "https://${cachefly_service.web_app.unique_name}.cachefly.net"
    api     = "https://${cachefly_service.api.unique_name}.cachefly.net"
    assets  = "https://${cachefly_service.assets.unique_name}.cachefly.net"
  }
}

output "verification" {
  description = "Verification that data sources work"
  value = {
    web_app_verified = data.cachefly_service.web_app_verify.id == cachefly_service.web_app.id
    api_verified     = data.cachefly_service.api_verify.id == cachefly_service.api.id
    all_services_active = alltrue([
      cachefly_service.web_app.status == "ACTIVE",
      cachefly_service.api.status == "ACTIVE",
      cachefly_service.assets.status == "ACTIVE"
    ])
  }
}