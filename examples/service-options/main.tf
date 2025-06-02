terraform {
  required_version = ">= 1.0"
  required_providers {
    cachefly = {
      source = "cachefly.com/avvvet/cachefly"
    }
  }
}

provider "cachefly" {
  api_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjQ0NDMxLCJ1c2VyIjoiNjgxYjNjZmIyNzE1MzEwMDM1Y2I3MmI2IiwidG9rZW4iOiI2ODNkMWU3YTc4M2NmOTAwNDA1YzdmZTQiLCJpYXQiOjE3NDg4MzU5NjJ9.IMvZq_jFCRoR4s8C63cLBmDvy1p-k80GOCWm3bF4o9M"
}

# Data source to fetch an existing service
data "cachefly_service" "example" {
  id = "681b3dc52715310035cb75d4"
}

# Enhanced service options with reverse proxy
resource "cachefly_service_options" "minimal" {
  service_id = data.cachefly_service.example.id
  
  # Basic boolean options
  cors                     = true
  brotli_compression      = true
  brotli_support          = true
  auto_redirect           = false
  serve_stale             = true
  
  # Basic caching options
  cache_by_geo_country    = false
  cache_by_region         = false
  normalize_query_string  = true
  
  # Request handling
  allow_retry             = true
  follow_redirect         = false
  send_xff                = true

  force_orig_qstring      = true
  
  # API defaults
  ftp                     = true  
  api_key_enabled         = true 
  
  error_ttl = {
    enabled               = true
    value                 = 400
  }

  con_timeout = {
    enabled               = true
    value                 = 5
  }

  max_cons = {
    enabled               = true
    value                 = 700
  }

  ttfb_timeout = {
    enabled               = true
    value                 = 7
  }

  shared_shield = {
  enabled = true
  value   = "IAD"  # Value must be one of: IAD, ORD, FRA, VIE
}

origin_hostheader = {
  enabled = true
  value   = ["example.com", "api.example.com"]
}

  # Reverse proxy configuration
  reverse_proxy = {
    enabled               = true
    hostname              = "backend.example.com"
    prepend               = "/api/v1"
    ttl                   = 3600
    cache_by_query_param  = true
    origin_scheme         = "FOLLOW"
    use_robots_txt        = true
    mode                  = "WEB"
  }
}

# Output the configured options
output "configured_options" {
  value = {
    # Basic options
    cors_enabled = cachefly_service_options.minimal.cors
    brotli_enabled = cachefly_service_options.minimal.brotli_compression
    serve_stale_enabled = cachefly_service_options.minimal.serve_stale
    normalize_query_string = cachefly_service_options.minimal.normalize_query_string
    
    # Reverse proxy info
    reverse_proxy_enabled = cachefly_service_options.minimal.reverse_proxy.enabled
    reverse_proxy_hostname = cachefly_service_options.minimal.reverse_proxy.hostname
    reverse_proxy_scheme = cachefly_service_options.minimal.reverse_proxy.origin_scheme
    
  }
}

# Output reverse proxy details
output "reverse_proxy_config" {
  value = cachefly_service_options.minimal.reverse_proxy
}

# Origin Error TTL
output "origin_error_ttl_option" {
  value = cachefly_service_options.minimal.error_ttl
}

# Origin ConTimeout
output "connection_timeout_option" {
  value = cachefly_service_options.minimal.con_timeout
}
   