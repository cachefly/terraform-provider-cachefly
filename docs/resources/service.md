---
page_title: "cachefly_service Resource - terraform-provider-cachefly"
subcategory: ""
description: |-
  Manages a CacheFly CDN service.
---

# cachefly_service (Resource)

Manages a CacheFly CDN service.

## Example Usage

```terraform
resource "cachefly_service" "website" {
  name               = "my-website"
  unique_name        = "my-website-prod"
  description        = "Main website CDN service"
  auto_ssl           = true
  configuration_mode = "API_RULES_AND_OPTIONS"
}
```

## Schema

### Required

- `name` (String) - The display name of the service
- `unique_name` (String) - A unique identifier for the service. Must be globally unique across all CacheFly accounts.

### Optional

- `description` (String) - A description of the service
- `auto_ssl` (Boolean) - Whether to automatically provision SSL certificates. Defaults to `false`
- `configuration_mode` (String) - Configuration mode. Valid values: `API_RULES_AND_OPTIONS`, `LEGACY`. Defaults to `API_RULES_AND_OPTIONS`

### Read-Only

- `id` (String) - The unique identifier of the service
- `created_at` (String) - When the service was created
- `updated_at` (String) - When the service was last updated
- `status` (String) - The current status of the service

## Import

Import is supported using the service ID:

```shell
terraform import cachefly_service.example service-id-here
```