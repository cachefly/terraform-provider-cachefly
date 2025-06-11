# ===================================================================
# CacheFly CDN - This example creates a full CDN service with domain and advanced options
# 
# ===================================================================

terraform {
  required_providers {
    cachefly = {
      source  = "cachefly.com/avvvet/cachefly" # todo: cachefly/cachefly
      version = "0.1.0"
    }
  }
}

provider "cachefly" {
  api_token = ""
}

# Variables for configuration
variable "cachefly_api_token" {
  description = "CacheFly API token"
  type        = string
  sensitive   = true
}

variable "base_domain" {
  description = "Base domain for the CDN service"
  type        = string
  default     = "ethobingo.online"
}

variable "company_name" {
  description = "Company name for service naming"
  type        = string
  default     = "YellowCompany"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  default     = "prod"
}

# Create the main CDN service
resource "cachefly_service" "web_app" {
  name               = "${var.company_name}-web-${var.environment}"
  unique_name        = "yellowcompany-web-prod-010"
  description        = "Web application CDN for ${var.company_name} ${var.environment} environment"
  auto_ssl           = true
  configuration_mode = "API_RULES_AND_OPTIONS"
}

# Main website domain
resource "cachefly_service_domain" "web_main" {
  service_id      = cachefly_service.web_app.id
  name            = var.base_domain
  description     = "Main website domain for ${var.company_name}"
  validation_mode = "DNS"
}

# Additional subdomain (optional)
resource "cachefly_service_domain" "web_cdn" {
  service_id      = cachefly_service.web_app.id
  name            = "cdn.${var.base_domain}"
  description     = "CDN subdomain for static assets"
  validation_mode = "DNS"
}

# API subdomain (optional)
resource "cachefly_service_domain" "web_api" {
  service_id      = cachefly_service.web_app.id
  name            = "api.${var.base_domain}"
  description     = "API subdomain with caching"
  validation_mode = "DNS"
}

# Advanced service options configuration
resource "cachefly_service_options" "advanced_options" {
  service_id = cachefly_service.web_app.id
  
  # Performance and compression options
  cors                    = true
  brotli_compression     = true
  brotli_support         = true
  auto_redirect          = false
  serve_stale            = true
  
  # Caching behavior
  cache_by_geo_country   = false
  cache_by_region        = false
  normalize_query_string = true
  
  # Request handling
  allow_retry            = true
  follow_redirect        = false
  send_xff               = true
  force_orig_qstring     = true
  
  # API and FTP settings
  ftp                    = true  
  api_key_enabled        = false 
  
  # Timeout and connection settings
  error_ttl = {
    enabled = true
    value   = 300  # 5 minutes for error caching
  }
  
  con_timeout = {
    enabled = true
    value   = 10   # 10 seconds connection timeout
  }
  
  max_cons = {
    enabled = true
    value   = 700  # Maximum concurrent connections
  }
  
  ttfb_timeout = {
    enabled = true
    value   = 15   # 15 seconds time to first byte
  }
  
  # Shield configuration for better performance
  shared_shield = {
    enabled = true
    value   = "IAD"  # Available options: IAD, ORD, FRA, VIE
  }
  
  # Origin host header configuration
  origin_hostheader = {
    enabled = true
    value   = [var.base_domain, "www.${var.base_domain}"]
  }


  # Reverse proxy configuration for API endpoints
  reverse_proxy = {
    enabled              = true
    hostname             = "backend.${var.base_domain}"
    prepend              = "/api/v1"
    ttl                  = 3600  # 1 hour cache TTL
    cache_by_query_param = true
    origin_scheme        = "FOLLOW"
    use_robots_txt       = true
    mode                 = "WEB"
  }
}

# Outputs
output "service_info" {
  description = "Information about the created CDN service"
  value = {
    service_id   = cachefly_service.web_app.id
    service_name = cachefly_service.web_app.name
    unique_name  = cachefly_service.web_app.unique_name
    auto_ssl     = cachefly_service.web_app.auto_ssl
  }
}

output "domains" {
  description = "Configured domains for the CDN service"
  value = {
    main_domain = cachefly_service_domain.web_main.name
    cdn_domain  = cachefly_service_domain.web_cdn.name
    api_domain  = cachefly_service_domain.web_api.name
  }
}

output "service_options_summary" {
  description = "Summary of configured service options"
  value = {
    cors_enabled           = cachefly_service_options.advanced_options.cors
    brotli_enabled        = cachefly_service_options.advanced_options.brotli_compression
    shared_shield_location = cachefly_service_options.advanced_options.shared_shield.value
    reverse_proxy_enabled = cachefly_service_options.advanced_options.reverse_proxy.enabled
  }
}

# Data source to verify the service after creation
data "cachefly_service" "created_service" {
  id = cachefly_service.web_app.id
  
  depends_on = [
    cachefly_service.web_app,
    cachefly_service_options.advanced_options
  ]
}

output "service_verification" {
  description = "Verification data for the created service"
  value = {
    service_exists = data.cachefly_service.created_service.id != null
    service_status = data.cachefly_service.created_service.status
  }
}