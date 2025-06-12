---
page_title: "cachefly_origin Resource - terraform-provider-cachefly"
subcategory: ""
description: |-
  Manages a CacheFly origin server.
---

# cachefly_origin (Resource)

Manages a CacheFly origin server that serves content to the CDN.

## Example Usage

```terraform
resource "cachefly_origin" "web_server" {
  type                       = "WEB"
  name                       = "main-web-server"
  host                       = "web.example.com"
  scheme                     = "HTTPS"
  cache_by_query_param       = false
  gzip                       = true
  ttl                        = 86400
  missed_ttl                 = 300
  connection_timeout         = 15
  time_to_first_byte_timeout = 15
}
```

## Schema

### Required

- `type` (String) - The type of origin. Valid values: `WEB`
- `name` (String) - The name of the origin
- `host` (String) - The hostname of the origin server

### Optional

- `scheme` (String) - The protocol scheme. Valid values: `HTTP`, `HTTPS`. Defaults to `HTTP`
- `cache_by_query_param` (Boolean) - Whether to cache by query parameters. Defaults to `false`
- `gzip` (Boolean) - Whether to enable gzip compression. Defaults to `false`
- `ttl` (Number) - Time to live in seconds for cached content. Defaults to `86400`
- `missed_ttl` (Number) - Time to live in seconds for cache misses. Defaults to `300`
- `connection_timeout` (Number) - Connection timeout in seconds. Defaults to `15`
- `time_to_first_byte_timeout` (Number) - Time to first byte timeout in seconds. Defaults to `15`

### Read-Only

- `id` (String) - The unique identifier of the origin
- `created_at` (String) - When the origin was created
- `updated_at` (String) - When the origin was last updated

## Import

Import is supported using the origin ID:

```shell
terraform import cachefly_origin.example origin-id-here
```