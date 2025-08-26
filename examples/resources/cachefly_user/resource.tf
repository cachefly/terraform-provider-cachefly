
# ===================================================================
# Example users resource administration
#
# Username can not be reused, even after deleted
# ===================================================================

terraform {
  required_providers {
    cachefly = {
      source  = "cachefly/cachefly" 
      version = "0.1.0"
    }
  }
}

provider "cachefly" {
  api_token = ""
}


resource "cachefly_user" "support_user" {
  username               = "ssemantha2.mark"
  email                  = "semantha.mark@example.com"
  full_name             = "Semantha Mark (account)"
  phone                 = "+1-555-123-4567"
  password              = "SecurePassword123!"
  password_change_required = true
  
  
  # Permissions to grant to the user
  permissions = [
    "P_ACCOUNT_VIEW",
    #"P_BILLING_VIEW",
    "P_SERVICE_PURGE",
    "P_SERVICE_MANAGE"
  ]
}

resource "cachefly_user" "billing_user" {
  username               = "ssam2.moller"
  email                  = "sam.moller@example.com"
  full_name             = "Sam Moller (billig)"
  phone                 = "+1-555-123-4567"
  password              = "SecurePassword123!"
  password_change_required = true
  
 
  
  # Permissions to grant to the user
  permissions = [
    "P_ACCOUNT_VIEW",
    "P_BILLING_VIEW",
    "P_SERVICE_PURGE",
    "P_SERVICE_MANAGE"
  ]
}

# created user verification 
output "support_user_id" {
  description = "The ID of the created user"
  value       = cachefly_user.support_user.id
}

output "support_user_status" {
  description = "Status of the created user"
  value       = cachefly_user.support_user.status
}

output "billing_user_id" {
  description = "The ID of the created user"
  value       = cachefly_user.billing_user.id
}

output "billing_user_status" {
  description = "Status of the created user"
  value       = cachefly_user.billing_user.status
}


