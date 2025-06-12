---
page_title: "cachefly_service Data Source - terraform-provider-cachefly"
subcategory: ""
description: |-
  Get information about a CacheFly service.
---

# cachefly_service (Data Source)

Get information about a CacheFly service by ID or unique name.

## Example Usage

```terraform
# Look up service by ID
data "cachefly_service" "by_id" {
  id = "service-id-here"
}

# Look up service by unique name
data "cachefly_service" "by_name" {
  unique_name = "my-service-unique-name"
}

# Look up service with additional options
data "cachefly_service" "detailed" {
  id               = "service-id-here"
  response_type    = "detailed"
  include_features = true
}

output "service_name" {
  value = data.cachefly_service.by_id.name
}
```

## Schema

### Optional

- `id` (String) - The unique identifier of the service. Either `id` or `unique_name` must be specified
- `unique_name` (String) - The unique name of the service. Either `id` or `unique_name` must be specified
- `response_type` (String) - The response type for the API call. Controls the level of detail returned
- `include_features` (Boolean) - Whether to include features in the response

### Read-Only

- `name` (String) - The display name of the service
- `auto_ssl` (Boolean) - Whether SSL certificates are automatically provisioned
- `configuration_mode` (String) - The configuration mode of the service
- `status` (String) - The current status of the service
- `created_at` (String) - When the service was created
- `updated_at` (String) - When the service was last updated