# ===================================================================
# Example certificate upload and service matching
# ===================================================================


terraform {
  required_providers {
    cachefly = {
      source  = "cachefly.com/avvvet/cachefly" # todo: cachefly/cachefly
      version = "0.1.0"
    }
  }
}

provider "cachefly" {
  api_token = ""
}


resource "cachefly_certificate" "web-east-2" {
  certificate = <<-EOT
-----BEGIN CERTIFICATE-----

-----END CERTIFICATE-----
  EOT

  certificate_key = <<-EOT
-----BEGIN PRIVATE KEY-----

-----END PRIVATE KEY-----
  EOT
}


resource "cachefly_service" "yellow_web_app" {
  name               = "yellow-web-app"
  unique_name        = "yellow-web-app-dev"
  description        = "Web application CDN for example.com"
  auto_ssl           = true  
  configuration_mode = "API_RULES_AND_OPTIONS"
}


resource "cachefly_service_domain" "app_main" {
  service_id      = cachefly_service.yellow_web_app.id
  name            = "example.com"  # This matches our certificate!
  description     = "Main website domain"
  validation_mode = "DNS"
}