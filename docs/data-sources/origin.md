---
page_title: "cachefly_origins Data Source - terraform-provider-cachefly"
subcategory: ""
description: |-
  Get information about CacheFly origins.
---

# cachefly_origins (Data Source)

Get information about CacheFly origins.

## Example Usage

```terraform
data "cachefly_origins" "web_origins" {
  type = "WEB"
}

output "origin_count" {
  value = length(data.cachefly_origins.web_origins.origins)
}
```

## Schema

### Optional

- `type` (String) - Filter origins by type. Valid values: `WEB`

### Read-Only

- `origins` (List) - List of origins matching the filter criteria
- `id` (String) - The data source identifier