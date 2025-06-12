---
page_title: "cachefly_service_domains Data Source - terraform-provider-cachefly"
subcategory: ""
description: |-
  Get information about domains for a CacheFly service.
---

# cachefly_service_domains (Data Source)

Get information about domains for a CacheFly service.

## Example Usage

```terraform
data "cachefly_service_domains" "website_domains" {
  service_id = cachefly_service.website.id
}

output "domain_count" {
  value = length(data.cachefly_service_domains.website_domains.domains)
}
```

## Schema

### Required

- `service_id` (String) - The ID of the service to get domains for

### Read-Only

- `domains` (List) - List of domains for the service
- `id` (String) - The data source identifier