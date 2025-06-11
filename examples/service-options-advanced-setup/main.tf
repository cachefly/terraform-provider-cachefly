# ===================================================================
# CacheFly CDN Advanced Service Options Configuration
# ===================================================================

terraform {
  required_version = ">= 1.0"
  required_providers {
    cachefly = {
      source = "cachefly.com/avvvet/cachefly"
    }
  }
}

provider "cachefly" {
  api_token = ""
}

# Data source to fetch an existing service
data "cachefly_service" "example" {
  id = "681b3dc52715310035cb75d4"
}

resource "cachefly_service_options" "comprehensive_options" {
  service_id = data.cachefly_service.example.id
  
  options = {
    # ===================================================================
    # BASIC OPTIONS
    # ===================================================================
    allowretry           = true
    forceorigqstring     = true
    nocache              = true
    servestale           = true
    normalizequerystring = true
    
    # ===================================================================
    # REQUEST HANDLING
    # ===================================================================
    "send-xff"       = true
    brotli_support   = true
    
    # ===================================================================
    # PURGE OPTIONS
    # ===================================================================
    purgenoquery = true
    
    # ===================================================================
    # REVERSE PROXY CONFIGURATION
    # ===================================================================
    reverseProxy = {
      enabled           = true
      mode              = "WEB"
      hostname          = "www.example.com"
      cacheByQueryParam = true
      originScheme      = "FOLLOW"
      ttl               = 2678400
      useRobotsTxt      = true
    }
    
    # ===================================================================
    # TIMEOUT & CONNECTION OPTIONS
    # ===================================================================
    error_ttl = {
      enabled = true
      value   = 700
    }
    
    ttfb_timeout = {
      enabled = true
      value   = 30
    }
    
    contimeout = {
      enabled = true
      value   = 10
    }
    
    maxcons = {
      enabled = true
      value   = 100
    }
    
    # ===================================================================
    # PERFORMANCE & DELIVERY OPTIONS
    # ===================================================================
    sharedshield = {
      enabled = true
      value   = "ORD"  # Chicago data center
    }
    
    purgemode = {
      enabled = true
      value   = "2"
    }
    
    bwthrottle = {
      enabled = true
      value   = 70656
    }
    
    # ===================================================================
    # ORIGIN & HEADERS
    # ===================================================================
    originhostheader = {
      enabled = true
      value   = ["origin.example.com", "backup.example.com"]
    }
    
    # ===================================================================
    # DIRECTORY & FILE HANDLING
    # ===================================================================
    dirpurgeskip = {
      enabled = true
      value   = 1
    }
    
    skip_encoding_ext = {
      enabled = true
      value   = [".zip", ".gz", ".tar", ".rar"]
    }
    
    skip_pserve_ext = {
      enabled = true
      value   = [".jpg", ".png", ".gif", ".css", ".js"]
    }
    
    # ===================================================================
    # GEOGRAPHIC & ADVANCED CACHING
    # ===================================================================
    cachebygeocountry = true
    cachebyreferer    = true
    cachebyregion     = true
    
    #expiryHeaders = []  to remove it

    expiryHeaders = [
      {
        path       = "/yellow"
        extension  = "test"
        expiryTime = 5
      },
      {
        path       = "/images"
        expiryTime = 3600
      },
      {
        extension  = ".jpg"
        expiryTime = 86400
      }
    ]
    
    # ===================================================================
    # CORS & REDIRECTS
    # ===================================================================
    cors         = true
    autoRedirect = true
    
    redirect = {
      enabled = true
      value   = "https://www.newdomain.com/"
    }
    
    # ===================================================================
    # ADVANCED FEATURES
    # ===================================================================
    livestreaming = true
    linkpreheat   = true
    
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
    protectServeKeyEnabled = true
    apiKeyEnabled          = true
    
    
  }
}

# Outputs
output "comprehensive_options" {
  value = cachefly_service_options.comprehensive_options.options
}