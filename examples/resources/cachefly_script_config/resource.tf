
# ===================================================================
# Example script config 
# ===================================================================

# ===================================================================
# ⚠️  IMPORTANT: Service Script Definition Constraint
# ===================================================================
# A service can have MULTIPLE different script configs attached to it,
# but the SAME script definition cannot be attached to the same service twice.
#
# Error: "Could not create script config, unexpected error: API error 400: 
# {"message":"Some services are already linked to another script config: 681b3dc52715310035cb75d4."}"
#
# SOLUTIONS:
# 1. Use different script config definitions for the same service
# 2. Update the existing script config instead of creating a new one
# 3. Remove/deactivate the existing script config with same definition first
# 4. Check existing script configs with: data "cachefly_script_configs" "existing" {}

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

# Create and configure URL redirects script config 
resource "cachefly_script_config" "url_redirects" {
  name                     = "url-redirects-dev-east"
  services                 = ["681b3dc52715310035cb75d4"]
  script_config_definition = "63fcfcc58a797a005f2ad04e"
  mime_type               = "text/json"
  activated               = true  
  
  value = jsonencode({
    "301" = {
      "/old/path/to/file.jpg"  = "https://www.sdk.com/path/to/new/file.jpg"
      "/old/path/to/file2.jpg" = "https://www.sdk.com/path/to/some/other/file.jpg"
    }
  })
}

# Create and configure AWS credentials script config
resource "cachefly_script_config" "aws_credentials" {
  name                     = "aws-credentials-dev-east"
  services                 = ["681b3dc52715310035cb75d4"]  
  script_config_definition = "643fea259be9a40060ba6298"
  mime_type               = "text/json"
  activated               = true
  
  value = jsonencode({
    aws_accessKey = "ACCESSKEYXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
    aws_secretKey = "SECRETKEYXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
    aws_region    = "us-west-1"  # Replace with actual region like us-east-1, us-west-2, etc.
    aws_version   = "v4"
  })
}



