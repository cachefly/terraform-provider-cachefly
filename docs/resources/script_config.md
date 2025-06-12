---
page_title: "cachefly_script_config Resource - terraform-provider-cachefly"
subcategory: ""
description: |-
  Manages a script configuration for CacheFly services.
---

# cachefly_script_config (Resource)

Manages a script configuration for CacheFly services to enable automation and custom logic.

## Example Usage

```terraform
resource "cachefly_script_config" "url_redirects" {
  name                     = "url-redirects"
  services                 = [cachefly_service.website.id]
  script_config_definition = "63fcfcc58a797a005f2ad04e"
  mime_type               = "text/json"
  activated               = true
  
  value = jsonencode({
    "301" = {
      "/old-page"    = "https://example.com/new-page"
      "/old-product" = "https://example.com/products/new-product"
    }
  })
}
```

## Schema

### Required

- `name` (String) - The name of the script configuration
- `services` (List of String) - List of service IDs to apply the script to
- `script_config_definition` (String) - The script configuration definition ID
- `mime_type` (String) - The MIME type of the configuration value
- `value` (String) - The configuration value (typically JSON)

### Optional

- `activated` (Boolean) - Whether the script configuration is activated. Defaults to `false`

### Read-Only

- `id` (String) - The unique identifier of the script configuration
- `created_at` (String) - When the script configuration was created
- `updated_at` (String) - When the script configuration was last updated

## Import

Import is supported using the script configuration ID:

```shell
terraform import cachefly_script_config.example script-config-id-here
```