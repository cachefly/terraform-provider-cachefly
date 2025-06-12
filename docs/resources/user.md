---
page_title: "cachefly_user Resource - terraform-provider-cachefly"
subcategory: ""
description: |-
  Manages a user account with permissions for CacheFly services.
---

# cachefly_user (Resource)

Manages a user account with permissions for CacheFly services.

## Example Usage

```terraform
resource "cachefly_user" "admin_user" {
  username                 = "admin.user"
  email                    = "admin@example.com"
  full_name               = "Admin User"
  phone                   = "+1-555-123-4567"
  password                = "SecurePassword123!"
  password_change_required = true
  
  services = [
    cachefly_service.website.id,
    cachefly_service.api.id
  ]
  
  permissions = [
    "P_ACCOUNT_VIEW",
    "P_BILLING_VIEW",
    "P_SERVICE_PURGE",
    "P_SERVICE_MANAGE"
  ]
}
```

## Schema

### Required

- `username` (String) - The username for the user account. Must be unique across all CacheFly.
- `email` (String) - The email address of the user
- `full_name` (String) - The full name of the user
- `password` (String, Sensitive) - The password for the user account

### Optional

- `phone` (String) - The phone number of the user
- `password_change_required` (Boolean) - Whether the user must change password on first login. Defaults to `false`
- `services` (List of String) - List of service IDs the user has access to
- `permissions` (List of String) - List of permissions granted to the user

### Read-Only

- `id` (String) - The unique identifier of the user
- `created_at` (String) - When the user was created
- `updated_at` (String) - When the user was last updated

## Import

Import is supported using the user ID:

```shell
terraform import cachefly_user.example user-id-here
```