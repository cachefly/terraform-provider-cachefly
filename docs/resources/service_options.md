---
page_title: "cachefly_service_options Resource - terraform-provider-cachefly"
subcategory: ""
description: |-
  Manages advanced configuration options for a CacheFly service.
---

# cachefly_service_options (Resource)

Manages advanced configuration options for a CacheFly service.

## Example Usage

```terraform
resource "cachefly_service_options" "website_options" {
  service_id = cachefly_service.website.id
  
  options = {
    allowretry           = true
    forceorigqstring     = false
    nocache              = false
    servestale           = true
    normalizequerystring = true
    brotli_support       = true
    "send-xff"           = true
    
    reverseProxy = {
      enabled           = true
      mode              = "WEB"
      hostname          = "web.example.com"
      cacheByQueryParam = false
      originScheme      = "HTTPS"
      ttl               = 86400
      useRobotsTxt      = true
    }
    
    error_ttl = {
      enabled = true
      value   = 300
    }
    
    cors         = true
    autoRedirect = true
  }
}
```

## Schema

### Required

- `service_id` (String) - The ID of the service to configure

### Optional

- `options` (Map) - Configuration options for the service. Available options include:
  - `allowretry` (Boolean) - Allow retry on origin failure
  - `forceorigqstring` (Boolean) - Force original query string
  - `nocache` (Boolean) - Disable caching
  - `servestale` (Boolean) - Serve stale content while updating
  - `normalizequerystring` (Boolean) - Normalize query string parameters
  - `brotli_support` (Boolean) - Enable Brotli compression
  - `send-xff` (Boolean) - Send X-Forwarded-For header
  - `cors` (Boolean) - Enable CORS
  - `autoRedirect` (Boolean) - Enable automatic redirects
  - `reverseProxy` (Object) - Reverse proxy configuration
  - `error_ttl` (Object) - Error TTL configuration
  - `ttfb_timeout` (Object) - Time to first byte timeout
  - `contimeout` (Object) - Connection timeout
  - `maxcons` (Object) - Maximum connections
  - `expiryHeaders` (List) - Cache expiry headers configuration
  - `httpmethods` (Object) - HTTP methods configuration

### Read-Only

- `id` (String) - The unique identifier of the service options
- `created_at` (String) - When the options were created
- `updated_at` (String) - When the options were last updated

## Import

Import is supported using the service ID:

```shell
terraform import cachefly_service_options.example service-id-here
```