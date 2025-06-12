---
page_title: "cachefly_service_options Data Source - terraform-provider-cachefly"
subcategory: ""
description: |-
  Get configuration options for a CacheFly service.
---

# cachefly_service_options (Data Source)

Get configuration options for a CacheFly service.

## Example Usage

```terraform
data "cachefly_service_options" "website_options" {
  service_id = cachefly_service.website.id
}

output "cors_enabled" {
  value = data.cachefly_service_options.website_options.cors
}

output "reverse_proxy_hostname" {
  value = data.cachefly_service_options.website_options.reverse_proxy.hostname
}
```

## Schema

### Required

- `service_id` (String) - The ID of the service to fetch options for

### Read-Only

- `ftp` (Boolean) - FTP access enabled for the service
- `cors` (Boolean) - CORS headers enabled
- `auto_redirect` (Boolean) - Automatic redirects enabled
- `brotli_support` (Boolean) - Brotli compression support enabled
- `nocache` (Boolean) - Caching disabled
- `cache_by_geo_country` (Boolean) - Cache by geographic country enabled
- `normalize_query_string` (Boolean) - Query string normalization enabled
- `allow_retry` (Boolean) - Retry on origin failures enabled
- `serve_stale` (Boolean) - Serve stale content when origin unavailable
- `send_xff` (Boolean) - Send X-Forwarded-For header enabled
- `protect_serve_key_enabled` (Boolean) - Protect serve key enabled
- `api_key_enabled` (Boolean) - API key authentication enabled
- `reverse_proxy` (Object) - Reverse proxy configuration with attributes:
  - `enabled` (Boolean) - Reverse proxy enabled
  - `hostname` (String) - Hostname for reverse proxy
  - `ttl` (Number) - TTL for reverse proxy cache
  - `cache_by_query_param` (Boolean) - Cache by query parameters
  - `origin_scheme` (String) - Origin scheme (http or https)
  - `mode` (String) - Reverse proxy mode
- `expiry_headers` (List) - Cache expiry headers configuration with attributes:
  - `path` (String) - Path pattern matched
  - `extension` (String) - File extension matched
  - `expiry_time` (Number) - Expiry time in seconds
- `error_ttl` (Object) - Error TTL configuration with attributes:
  - `enabled` (Boolean) - Error TTL enabled
  - `value` (String) - Error TTL value
- `ttfb_timeout` (Object) - Time to first byte timeout configuration
- `shared_shield` (Object) - Shared shield configuration
- `http_methods` (Object) - HTTP methods configuration