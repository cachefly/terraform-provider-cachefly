# ===================================================================
# CacheFly CDN - Complete Feature Showcase
# A comprehensive example demonstrating available CDN features
# ===================================================================

terraform {
  required_version = ">= 1.0"
  required_providers {
    cachefly = {
      source  = "cachefly/cachefly"
      version = "0.1.0"
    }
  }
}

provider "cachefly" {
  api_token = ""  # Add your CacheFly API token here
}

# ===================================================================
# VARIABLES - Customize for your setup
# ===================================================================

variable "domain" {
  description = "Your main domain"
  type        = string
  default     = "myuniquedomain123.com"  # Change this to your actual domain
}

variable "project_name" {
  description = "Name of your project"
  type        = string
  default     = "my-website"
}

variable "unique_suffix" {
  description = "Unique suffix to avoid naming conflicts (change for each run)"
  type        = string
  default     = "v8"  # Change this for each deployment (v1, v2, v3, etc.)
}

variable "admin_username" {
  description = "Admin user username (must be unique across all CacheFly)"
  type        = string
  default     = "admin.user.v8"  # Change this for each deployment
}

variable "support_username" {
  description = "Support user username (must be unique across all CacheFly)"
  type        = string
  default     = "support.user.v8"  # Change this for each deployment
}

# ===================================================================
# ORIGINS - Backend servers that serve your content
# ===================================================================

resource "cachefly_origin" "web_server" {
  type                       = "WEB"
  name                       = "main-web-server-${var.unique_suffix}"
  host                       = "web.${var.domain}"
  scheme                     = "HTTPS"
  cache_by_query_param       = false
  gzip                       = true
  ttl                        = 86400  # 24 hours
  missed_ttl                 = 300    # 5 minutes
  connection_timeout         = 15
  time_to_first_byte_timeout = 15
}

resource "cachefly_origin" "api_server" {
  type                 = "WEB"
  name                 = "api-server-${var.unique_suffix}"
  host                 = "api.${var.domain}"
  scheme               = "HTTPS"
  cache_by_query_param = true  # Cache API responses by query params
  gzip                 = true
  ttl                  = 3600  # 1 hour
  missed_ttl           = 60    # 1 minute
}

# ===================================================================
# SERVICES - CDN services that deliver your content
# ===================================================================

resource "cachefly_service" "website" {
  name               = "${var.project_name}-website-${var.unique_suffix}"
  unique_name        = "${var.project_name}-website-${var.unique_suffix}"
  description        = "Main website CDN service"
  auto_ssl           = true
  configuration_mode = "API_RULES_AND_OPTIONS"
}

resource "cachefly_service" "api" {
  name               = "${var.project_name}-api-${var.unique_suffix}"
  unique_name        = "${var.project_name}-api-${var.unique_suffix}"
  description        = "API CDN service"
  auto_ssl           = true
  configuration_mode = "API_RULES_AND_OPTIONS"
}

resource "cachefly_service" "assets" {
  name               = "${var.project_name}-assets-${var.unique_suffix}"
  unique_name        = "${var.project_name}-assets-${var.unique_suffix}"
  description        = "Static assets CDN service"
  auto_ssl           = true
  configuration_mode = "API_RULES_AND_OPTIONS"
}

# ===================================================================
# DOMAINS - Custom domains for your services
# ===================================================================

resource "cachefly_service_domain" "main_domain" {
  service_id       = cachefly_service.website.id
  name             = var.domain
  description      = "Main website domain"
  validation_mode  = "DNS"
}

resource "cachefly_service_domain" "www_domain" {
  service_id       = cachefly_service.website.id
  name             = "www.${var.domain}"
  description      = "WWW subdomain"
  validation_mode  = "DNS"
}

resource "cachefly_service_domain" "api_domain" {
  service_id       = cachefly_service.api.id
  name             = "api.${var.domain}"
  description      = "API subdomain"
  validation_mode  = "DNS"
}

resource "cachefly_service_domain" "cdn_domain" {
  service_id       = cachefly_service.assets.id
  name             = "cdn.${var.domain}"
  description      = "CDN subdomain for static assets"
  validation_mode  = "DNS"
}

# ===================================================================
# SSL CERTIFICATES - Custom SSL certificates (optional)
# Commented out until you have valid certificates
# ===================================================================

# Uncomment and add your actual certificate when ready
# resource "cachefly_certificate" "main_cert" {
#   certificate = <<-EOT
# -----BEGIN CERTIFICATE-----
# MIIDXTCCAkWgAwIBAgIJAKoK/heBjcOuMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
# ... your actual certificate content here ...
# -----END CERTIFICATE-----
#   EOT
#   
#   certificate_key = <<-EOT
# -----BEGIN PRIVATE KEY-----
# MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDdwJmuFqW7RMuD
# ... your actual private key content here ...
# -----END PRIVATE KEY-----
#   EOT
# }

# ===================================================================
# SERVICE OPTIONS - Advanced CDN configuration
# ===================================================================

resource "cachefly_service_options" "website_options" {
  service_id = cachefly_service.website.id
  
  options = {
    # ===================================================================
    # BASIC CACHING OPTIONS
    # ===================================================================
    allowretry           = true
    forceorigqstring     = false  # Don't force original query strings
    nocache              = false  # Enable caching (set to true to disable)
    servestale           = true   # Serve stale content while updating
    normalizequerystring = true   # Normalize query string parameters
    purgenoquery         = true   # Purge without query parameters
    
    # ===================================================================
    # PERFORMANCE & COMPRESSION
    # ===================================================================
    brotli_support = true         # Enable Brotli compression
    "send-xff"     = true         # Send X-Forwarded-For header
    
    # ===================================================================
    # REVERSE PROXY CONFIGURATION
    # ===================================================================
    reverseProxy = {
      enabled           = true
      mode              = "WEB"
      hostname          = "web.${var.domain}"
      cacheByQueryParam = false
      originScheme      = "HTTPS"
      ttl               = 86400
      useRobotsTxt      = true
    }
    
    # ===================================================================
    # TIMEOUT & CONNECTION SETTINGS
    # ===================================================================
    error_ttl = {
      enabled = true
      value   = 300               # Cache errors for 5 minutes
    }
    
    ttfb_timeout = {
      enabled = true
      value   = 30                # Time to first byte timeout
    }
    
    contimeout = {
      enabled = true
      value   = 10                # Connection timeout
    }
    
    maxcons = {
      enabled = true
      value   = 100               # Maximum concurrent connections
    }
    
    # ===================================================================
    # PERFORMANCE & DELIVERY OPTIONS
    # ===================================================================
    sharedshield = {
      enabled = true
      value   = "ORD"             # Chicago data center for shield
    }
    
    purgemode = {
      enabled = true
      value   = "2"               # Purge mode setting
    }
    
    bwthrottle = {
      enabled = true
      value   = 70656             # Bandwidth throttling in KB/s
    }
    
    # ===================================================================
    # ORIGIN & HEADERS CONFIGURATION
    # ===================================================================
    originhostheader = {
      enabled = true
      value   = ["web.${var.domain}", "backup.${var.domain}"]
    }
    
    # ===================================================================
    # DIRECTORY & FILE HANDLING
    # ===================================================================
    dirpurgeskip = {
      enabled = true
      value   = 1                 # Skip directory purge
    }
    
    skip_encoding_ext = {
      enabled = true
      value   = [".zip", ".gz", ".tar", ".rar", ".7z"]  # Skip encoding for compressed files
    }
    
    skip_pserve_ext = {
      enabled = true
      value   = [".jpg", ".png", ".gif", ".css", ".js", ".ico", ".svg"]  # Skip processing for static assets
    }
    
    # ===================================================================
    # GEOGRAPHIC & ADVANCED CACHING
    # ===================================================================
    cachebygeocountry = true      # Cache by geographic country
    cachebyreferer    = true      # Cache by referer header
    cachebyregion     = true      # Cache by geographic region
    
    # ===================================================================
    # CACHE EXPIRY HEADERS
    # ===================================================================
    expiryHeaders = [
      {
        path       = "/images"
        expiryTime = 86400        # 24 hours for images
      },
      {
        path       = "/static"
        expiryTime = 604800       # 7 days for static content
      },
      {
        extension  = ".css"
        expiryTime = 3600         # 1 hour for CSS
      },
      {
        extension  = ".js"
        expiryTime = 3600         # 1 hour for JavaScript
      },
      {
        extension  = ".jpg"
        expiryTime = 86400        # 24 hours for JPEG images
      },
      {
        extension  = ".png"
        expiryTime = 86400        # 24 hours for PNG images
      },
      {
        extension  = ".pdf"
        expiryTime = 604800       # 7 days for PDF files
      }
    ]
    
    # ===================================================================
    # CORS & REDIRECTS
    # ===================================================================
    cors         = true           # Enable CORS
    autoRedirect = true           # Enable automatic redirects
    
    redirect = {
      enabled = false             # Set to true and add URL for redirects
      # value   = "https://www.newdomain.com/"
    }
    
    # ===================================================================
    # ADVANCED FEATURES
    # ===================================================================
    livestreaming = false         # Enable for live streaming content
    linkpreheat   = true          # Enable link preheating
    
    # ===================================================================
    # HTTP METHODS CONFIGURATION
    # ===================================================================
    httpmethods = {
      enabled = true
      value = {
        GET     = true
        POST    = true
        PUT     = false
        DELETE  = false
        HEAD    = true
        OPTIONS = true
        PATCH   = false
      }
    }
    
    # ===================================================================
    # SECURITY OPTIONS
    # ===================================================================
    protectServeKeyEnabled = true  # Enable serve key protection
    apiKeyEnabled          = true  # Enable API key authentication
  }
}

resource "cachefly_service_options" "api_options" {
  service_id = cachefly_service.api.id
  
  options = {
    # ===================================================================
    # API-SPECIFIC CACHING OPTIONS
    # ===================================================================
    allowretry           = true
    forceorigqstring     = true   # Force original query strings for APIs
    nocache              = false  # Enable caching for API responses
    normalizequerystring = false  # Preserve exact query strings for APIs
    purgenoquery         = false  # Include query parameters in purge
    
    # ===================================================================
    # PERFORMANCE OPTIONS FOR APIs
    # ===================================================================
    brotli_support = true         # Enable compression for API responses
    "send-xff"     = true         # Send client IP to origin
    
    # ===================================================================
    # REVERSE PROXY FOR API
    # ===================================================================
    reverseProxy = {
      enabled           = true
      mode              = "WEB"
      hostname          = "api.${var.domain}"
      cacheByQueryParam = true    # Cache API responses by query params
      originScheme      = "HTTPS"
      ttl               = 3600    # Shorter TTL for dynamic API content
      useRobotsTxt      = false   # APIs don't need robots.txt
    }
    
    # ===================================================================
    # API TIMEOUT SETTINGS
    # ===================================================================
    error_ttl = {
      enabled = true
      value   = 60              # Cache API errors for 1 minute only
    }
    
    ttfb_timeout = {
      enabled = true
      value   = 15              # Shorter timeout for APIs
    }
    
    contimeout = {
      enabled = true
      value   = 5               # Quick connection timeout for APIs
    }
    
    maxcons = {
      enabled = true
      value   = 200             # Higher concurrent connections for APIs
    }
    
    # ===================================================================
    # API CACHING BY LOCATION/USER
    # ===================================================================
    cachebygeocountry = false     # Usually not needed for APIs
    cachebyreferer    = false     # APIs don't typically cache by referer
    cachebyregion     = false     # Skip regional caching for APIs
    
    # ===================================================================
    # API CACHE EXPIRY (shorter times)
    # ===================================================================
    expiryHeaders = [
      {
        path       = "/api/v1/users"
        expiryTime = 300          # 5 minutes for user data
      },
      {
        path       = "/api/v1/static"
        expiryTime = 3600         # 1 hour for static API data
      },
      {
        path       = "/api/v1/realtime"
        expiryTime = 60           # 1 minute for real-time data
      }
    ]
    
    # ===================================================================
    # CORS FOR API ACCESS
    # ===================================================================
    cors         = true           # Essential for web API access
    autoRedirect = false          # APIs shouldn't auto-redirect
    
    # ===================================================================
    # HTTP METHODS FOR API
    # ===================================================================
    httpmethods = {
      enabled = true
      value = {
        GET     = true            # Read operations
        POST    = true            # Create operations
        PUT     = true            # Update operations
        DELETE  = true            # Delete operations
        HEAD    = true            # Metadata requests
        OPTIONS = true            # CORS preflight
        PATCH   = true            # Partial updates
      }
    }
    
    # ===================================================================
    # API SECURITY
    # ===================================================================
    protectServeKeyEnabled = true  # Protect API endpoints
    apiKeyEnabled          = true  # Require API keys
  }
}

# ===================================================================
# SCRIPT CONFIGURATIONS - Automation and custom logic
# ===================================================================

resource "cachefly_script_config" "url_redirects" {
  name                     = "url-redirects-${var.unique_suffix}"
  services                 = [cachefly_service.website.id]
  script_config_definition = "63fcfcc58a797a005f2ad04e"  # URL redirects script
  mime_type               = "text/json"
  activated               = true
  
  value = jsonencode({
    "301" = {
      "/old-page"     = "https://${var.domain}/new-page"
      "/old-product"  = "https://${var.domain}/products/new-product"
    }
  })
}

resource "cachefly_script_config" "aws_credentials" {
  name                     = "aws-credentials-${var.unique_suffix}"
  services                 = [cachefly_service.assets.id]
  script_config_definition = "643fea259be9a40060ba6298"  # AWS credentials script
  mime_type               = "text/json"
  activated               = true
  
  value = jsonencode({
    aws_accessKey = "YOUR_AWS_ACCESS_KEY_HERE"
    aws_secretKey = "YOUR_AWS_SECRET_KEY_HERE"
    aws_region    = "us-east-1"
    aws_version   = "v4"
  })
}

# ===================================================================
# USER MANAGEMENT - Team access control
# ===================================================================

resource "cachefly_user" "admin_user" {
  username                 = var.admin_username
  email                    = "admin@${var.domain}"
  full_name               = "Admin User"
  phone                   = "+1-555-123-4567"
  password                = "SecurePassword123!"
  password_change_required = true
  
  services = [
    cachefly_service.website.id,
    cachefly_service.api.id,
    cachefly_service.assets.id
  ]
  
  permissions = [
    "P_ACCOUNT_VIEW",
    "P_BILLING_VIEW",
    "P_SERVICE_PURGE",
    "P_SERVICE_MANAGE"
  ]
}

resource "cachefly_user" "support_user" {
  username                 = var.support_username
  email                    = "support@${var.domain}"
  full_name               = "Support User"
  phone                   = "+1-555-123-4568"
  password                = "SecurePassword123!"
  password_change_required = true
  
  services = [
    cachefly_service.website.id
  ]
  
  permissions = [
    "P_ACCOUNT_VIEW",
    "P_SERVICE_PURGE"
  ]
}

# ===================================================================
# DATA SOURCES - Minimal verification only
# ===================================================================

# Commented out data sources that might be causing API issues
# data "cachefly_origins" "all_origins" {
#   type = "WEB"
# }

# data "cachefly_service_domains" "website_domains" {
#   service_id = cachefly_service.website.id
# }

# ===================================================================
# OUTPUTS - Minimal essential information
# ===================================================================



output "setup_status" {
  description = "Basic setup verification"
  value = {
    services_created = "✅ 3 services created"
    domains_created  = "✅ 4 domains created"
    users_created    = "✅ 2 users created"
  }
}