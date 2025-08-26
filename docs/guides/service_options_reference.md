---
page_title: "Service options reference"
---


# Service Options Reference

This page lists the supported `options` for the `cachefly_service` resource and how to configure them in Terraform.

Two kinds of options exist:

- Standard options: simple booleans toggles.
- Dynamic options: structured objects that usually follow the pattern `{ enabled = <bool>, value = <typed value> }`. Some dynamic options define their own object schema instead of `value`.

See also: `../resources/service.md` → attribute `options`.

## Standard options

- protectServeKeyEnabled (Boolean)
  - Enables ProtectServe. When set to true, the provider regenerates the ProtectServe key; when set to false, the key is deleted.
  - Example:

```hcl
options = {
  protectServeKeyEnabled = true
}
```

- cors (Boolean)
  - Enables CORS override for the service.
  - Example:

```hcl
options = {
  cors = true
}
```

- referrerBlocking (Boolean)
  - Blocks requests based on referrer rules.
  - Example:

```hcl
options = {
  referrerBlocking = true
}
```

- autoRedirect (Boolean)
  - Automatically redirects HTTP to HTTPS.
  - Example:

```hcl
options = {
  autoRedirect = true
}
```

## Dynamic options

Some options are dynamic and use an object schema. The most common is `reverseProxy`.

### reverseProxy (Object)

Reverse proxy configuration. When `enabled = true`, the following fields are required unless stated otherwise.

- enabled (Boolean) — required
- hostname (String) — required
- originScheme (String) — required; one of: FOLLOW, HTTP, HTTPS
- ttl (Number) — required; cache TTL in seconds
- useRobotsTxt (Boolean) — required
- cacheByQueryParam (Boolean) — required
- mode (String) — optional; one of: WEB, OBJECT_STORAGE
  - If `mode = OBJECT_STORAGE`, the following are also required:
    - accessKey (String)
    - secretKey (String)
    - region (String)

Examples:

WEB origin

```hcl
options = {
  reverseProxy = {
    enabled           = true
    mode              = "WEB"
    hostname          = "origin.example.com"
    originScheme      = "HTTPS"    # FOLLOW | HTTP | HTTPS
    ttl               = 3600
    useRobotsTxt      = false
    cacheByQueryParam = true
  }
}
```

Object Storage origin

```hcl
options = {
  reverseProxy = {
    enabled           = true
    mode              = "OBJECT_STORAGE"
    hostname          = "bucket.example.com"
    originScheme      = "HTTPS"
    ttl               = 3600
    useRobotsTxt      = false
    cacheByQueryParam = true

    accessKey = var.object_storage_access_key
    secretKey = var.object_storage_secret_key
    region    = "us-east-1"
  }
}
```

## Notes and mappings

- API/UI naming vs Terraform keys (for reference):
  - Reverse Proxy → `reverseProxy`
  - ProtectServe → `protectServeKeyEnabled`
  - CORS Override → `cors`
  - Referrer Blocking → `referrerBlocking`
  - Auto HTTPS Redirect → `autoRedirect`
  - Expiry Overrides → `expiryHeaders` (may not be available for every service)

- The available options can vary by service and account. If an option is unsupported for a given service, the provider will report a validation error during apply.

- Dynamic options may introduce additional fields or constraints over time. When in doubt, prefer the examples above and consult the provider changelog.
