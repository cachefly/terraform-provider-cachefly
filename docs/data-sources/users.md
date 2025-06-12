---
page_title: "cachefly_users Data Source - terraform-provider-cachefly"
subcategory: ""
description: |-
  Get information about CacheFly users.
---

# cachefly_users (Data Source)

Get information about CacheFly users with optional filtering and pagination.

## Example Usage

```terraform
# Get all users
data "cachefly_users" "all_users" {}

# Search for specific users
data "cachefly_users" "admin_users" {
  search = "admin"
  limit  = 10
}

# Get users with pagination
data "cachefly_users" "page_two" {
  offset = 20
  limit  = 10
}

output "user_count" {
  value = data.cachefly_users.all_users.total_count
}

output "admin_emails" {
  value = [for user in data.cachefly_users.admin_users.users : user.email]
}
```

## Schema

### Optional

- `search` (String) - Search term to filter users by username, email, or full name
- `offset` (Number) - Number of users to skip for pagination
- `limit` (Number) - Maximum number of users to return
- `response_type` (String) - Response type for the API request

### Read-Only

- `id` (String) - The data source identifier
- `total_count` (Number) - Total number of users available (before pagination)
- `users` (List) - List of users matching the criteria with attributes:
  - `id` (String) - The unique identifier of the user
  - `username` (String) - Username of the user
  - `email` (String) - Email address of the user
  - `full_name` (String) - Full name of the user
  - `phone` (String) - Phone number of the user
  - `password_change_required` (Boolean) - Whether the user must change password on next login
  - `services` (Set of String) - Set of service IDs the user has access to
  - `permissions` (Set of String) - Set of permissions granted to the user
  - `status` (String) - Status of the user account
  - `created_at` (String) - When the user was created
  - `updated_at` (String) - When the user was last updated