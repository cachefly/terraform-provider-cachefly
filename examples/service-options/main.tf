terraform {
  required_version = ">= 1.0"
  required_providers {
    cachefly = {
      source = "cachefly.com/avvvet/cachefly"
    }
  }
}

provider "cachefly" {
  api_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjQ0NDMxLCJ1c2VyIjoiNjgxYjNjZmIyNzE1MzEwMDM1Y2I3MmI2IiwidG9rZW4iOiI2ODNkYWI5NDc4M2NmOTAwNDA1ZDY3OTciLCJpYXQiOjE3NDg4NzQ1MDd9.BfqNO3YEepe4T44GJPV2PZ-EZcz7B-loE6QVBWMDZaY"
}

# Data source to fetch an existing service
data "cachefly_service" "example" {
  id = "681b3dc52715310035cb75d4"
}

resource "cachefly_service_options" "advanced_options" {
  service_id = data.cachefly_service.example.id
  
  options = {
    nocache              = false
    allowretry           = true
    servestale           = true
    normalizequerystring = true
    forceorigqstring     = false

    reverseProxy = {
      enabled           = true
      mode              = "WEB"
      hostname          = "www.example.com"
      cacheByQueryParam = true
      originScheme      = "FOLLOW"
      ttl               = 2678400
      useRobotsTxt      = true
    }

    rawLogs = {
      enabled     = true
      logFormat   = "combined"
      compression = "gzip"
    }

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

    bwthrottle = {
      enabled = true
      value   = 1000000
    }

    sharedshield = {
      enabled = true
      value   = "ORD"
    }

    purgemode = {
      enabled = true
      value   = "2"
    }

    redirect = {
      enabled = true
      value   = "https://www.newdomain.com/"
    }

    slice = {
      enabled = true
      value   = true
    }

    originhostheader = {
      enabled = true
      value   = ["origin.example.com", "backup.example.com"]
    }

    skip_pserve_ext = {
      enabled = true
      value   = [".jpg", ".png", ".gif", ".css", ".js"]
    }

    skip_encoding_ext = {
      enabled = true
      value   = [".zip", ".gz", ".tar", ".rar"]
    }

    bwthrottlequery = {
      enabled = true
      value   = ["limit", "throttle"]
    }

    dirpurgeskip = {
      enabled = true
      value   = 1
    }

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
  }
}

# Outputs
output "advanced_options" {
  value = cachefly_service_options.advanced_options.options
}