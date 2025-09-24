# CacheFly Origins Example
# This example shows how to create and manage web origin servers

terraform {
  required_version = ">= 1.0"
  required_providers {
    cachefly = {
      source = "cachefly/cachefly" 
    }
  }
}

provider "cachefly" {
  api_token = ""
}

# ===================================================================
# WEB ORIGINS - Standard web servers
# ===================================================================

# Main web server origin
resource "cachefly_origin" "web_server" {
  type                       = "WEB"
  name                       = "main-web-server"
  hostname                   = "web.example.com"
  scheme                     = "HTTPS"
  cache_by_query_param       = false
  gzip                       = true
  ttl                        = 86400  # 24 hours
  missed_ttl                 = 300    # 5 minutes
  connection_timeout         = 15     # 15 seconds
  time_to_first_byte_timeout = 15     # 15 seconds
}

# API server origin
resource "cachefly_origin" "api_server" {
  type                 = "WEB"
  name                 = "api-server"
  hostname             = "api.example.com"
  scheme               = "HTTPS"
  cache_by_query_param = true  # Cache API responses by query params
  gzip                 = true
  ttl                  = 3600  # 1 hour
  missed_ttl           = 60    # 1 minute
}

# Development server with different settings
resource "cachefly_origin" "dev_server" {
  type                 = "WEB"
  name                 = "dev-server"
  hostname             = "dev.example.com"
  scheme               = "HTTPS"
  cache_by_query_param = false
  gzip                 = false  # Disable compression for development
  ttl                  = 300    # 5 minutes - short cache for dev
  missed_ttl           = 60     # 1 minute
}

# ===================================================================
# DATA SOURCES - Lookup and Verification
# ===================================================================

# List all origins
data "cachefly_origins" "all_origins" {
  # No filter - get all origins
}

# List only WEB origins
data "cachefly_origins" "web_origins" {
  type = "WEB"
}

# ===================================================================
# OUTPUTS - Useful Information
# ===================================================================

output "origins_summary" {
  description = "Summary of all created origins"
  value = {
    web_server = {
      id       = cachefly_origin.web_server.id
      name     = cachefly_origin.web_server.name
      hostname = cachefly_origin.web_server.hostname
      type     = cachefly_origin.web_server.type
      ttl      = cachefly_origin.web_server.ttl
    }
    api_server = {
      id       = cachefly_origin.api_server.id
      name     = cachefly_origin.api_server.name
      hostname = cachefly_origin.api_server.hostname
      type     = cachefly_origin.api_server.type
      ttl      = cachefly_origin.api_server.ttl
    }
    dev_server = {
      id       = cachefly_origin.dev_server.id
      name     = cachefly_origin.dev_server.name
      hostname = cachefly_origin.dev_server.hostname
      type     = cachefly_origin.dev_server.type
      ttl      = cachefly_origin.dev_server.ttl
    }
  }
}

output "origin_types" {
  description = "Different types of origins configured"
  value = {
    web_origins = [
      cachefly_origin.web_server.name,
      cachefly_origin.api_server.name,
      cachefly_origin.dev_server.name,
    ]
  }
}

output "cache_settings" {
  description = "Cache settings for each origin"
  value = {
    web_server = {
      ttl                  = cachefly_origin.web_server.ttl
      missed_ttl           = cachefly_origin.web_server.missed_ttl
      cache_by_query_param = cachefly_origin.web_server.cache_by_query_param
      gzip                 = cachefly_origin.web_server.gzip
    }
    api_server = {
      ttl                  = cachefly_origin.api_server.ttl
      missed_ttl           = cachefly_origin.api_server.missed_ttl
      cache_by_query_param = cachefly_origin.api_server.cache_by_query_param
      gzip                 = cachefly_origin.api_server.gzip
    }
    dev_server = {
      ttl                  = cachefly_origin.dev_server.ttl
      missed_ttl           = cachefly_origin.dev_server.missed_ttl
      cache_by_query_param = cachefly_origin.dev_server.cache_by_query_param
      gzip                 = cachefly_origin.dev_server.gzip
    }
  }
}

output "verification" {
  description = "Verification that everything is working"
  value = {
    total_origins        = length(data.cachefly_origins.all_origins.origins)
    web_origins_count    = length(data.cachefly_origins.web_origins.origins)
    all_origins_created  = length(data.cachefly_origins.all_origins.origins) >= 3
  }
}

output "next_steps" {
  description = "What to do next with your origins"
  value = {
    message = "🎉 Web origins created successfully! Next steps:"
    steps = [
      "1. Create services that will use these origins",
      "2. Attach origins to services via rules or configuration",
      "3. Set up custom domains for your services",
      "4. Configure caching rules and optimization settings",
      "5. Test CDN performance with your origin servers"
    ]
    note = "Origins are backend servers - you need to attach them to services to serve content"
    origin_types = "Using WEB origins for standard HTTP/HTTPS servers"
  }
}