---
page_title: "cachefly_service_domain Resource - terraform-provider-cachefly"
subcategory: ""
description: |-
  Manages a custom domain for a CacheFly service.
---

# cachefly_service_domain (Resource)

Manages a custom domain for a CacheFly service.

## Example Usage

```terraform
resource "cachefly_service_domain" "main_domain" {
  service_id      = cachefly_service.website.id
  name            = "example.com"
  description     = "Main website domain"
  validation_mode = "DNS"
}
```

## Schema

### Required

- `service_id` (String) - The ID of the service to attach the domain to
- `name` (String) - The domain name

### Optional

- `description` (String) - A description of the domain
- `validation_mode` (String) - The domain validation mode. Valid values: `DNS`. Defaults to `DNS`

### Read-Only

- `id` (String) - The unique identifier of the domain
- `created_at` (String) - When the domain was created
- `updated_at` (String) - When the domain was last updated
- `status` (String) - The current status of the domain

## Import

Import is supported using the domain ID:

```shell
terraform import cachefly_service_domain.example domain-id-here
```