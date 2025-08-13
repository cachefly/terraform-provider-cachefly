# ===================================================================
# CacheFly CDN - Log Target Setup Examples (S3, Elasticsearch, GCS)
# ===================================================================

terraform {
  required_version = ">= 1.0"
  required_providers {
    cachefly = {
      source = "cachefly/cachefly"
    }
  }
}

provider "cachefly" {
  # Uses CACHEFLY_API_TOKEN if set; otherwise, set api_token here
  # api_token = ""
}

# -------------------------------------------------------------------
# Example 1: Create a simple service to attach logs to (optional)
# -------------------------------------------------------------------
resource "cachefly_service" "example" {
  name        = "example-dev-svc"
  unique_name = "example-dev-svc-logs"
  description = "Service used to demonstrate log target attachment"
  auto_ssl    = true
}

# -------------------------------------------------------------------
# Example 2: S3 Bucket Log Target
# -------------------------------------------------------------------
resource "cachefly_log_target" "s3_logs" {
  name              = "example-dev-s3-logs"
  type              = "S3_BUCKET"
  bucket            = "my-log-bucket"
  region            = "us-east-1"
  signature_version = "v4"

  # Demo credentials; replace with real values or use TF variables/secrets
  access_key = "AKIAIOSFODNN7EXAMPLE"
  secret_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

  # Enable logs for selected services
  access_logs_services = [cachefly_service.example.id]
  origin_logs_services = [cachefly_service.example.id]

  depends_on = [cachefly_service.example]
}

# -------------------------------------------------------------------
# Example 3: Elasticsearch Log Target
# -------------------------------------------------------------------
resource "cachefly_log_target" "elasticsearch_logs" {
  name                         = "example-dev-es-logs"
  type                         = "ELASTICSEARCH"
  hosts                        = [
    "elasticsearch1.example.com:9200",
    "elasticsearch2.example.com:9200",
  ]
  ssl                          = true
  ssl_certificate_verification = true
  index                        = "cachefly-logs"
  user                         = "elastic"
  password                     = "replace-with-real-password"

  access_logs_services = [cachefly_service.example.id]
  origin_logs_services = [cachefly_service.example.id]

  depends_on = [cachefly_service.example]
}

# -------------------------------------------------------------------
# Example 4: Google Cloud Storage Log Target (uncomment to use)
# -------------------------------------------------------------------
# resource "cachefly_log_target" "gcs_logs" {
#   name     = "example-dev-gcs-logs"
#   type     = "GOOGLE_BUCKET"
#   bucket   = "my-gcp-log-bucket"
#   json_key = jsonencode({
#     type                        = "service_account"
#     project_id                  = "my-project-12345"
#     private_key_id              = "key-id-12345"
#     private_key                 = "-----BEGIN PRIVATE KEY-----\nMII...\n-----END PRIVATE KEY-----\n"
#     client_email                = "svc-account@my-project-12345.iam.gserviceaccount.com"
#     client_id                   = "123456789012345678901"
#     auth_uri                    = "https://accounts.google.com/o/oauth2/auth"
#     token_uri                   = "https://oauth2.googleapis.com/token"
#     auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
#     client_x509_cert_url        = "https://www.googleapis.com/robot/v1/metadata/x509/svc-account%40my-project-12345.iam.gserviceaccount.com"
#   })
# }

# -------------------------------------------------------------------
# Outputs
# -------------------------------------------------------------------
output "log_targets" {
  value = {
    s3 = {
      id        = cachefly_log_target.s3_logs.id
      name      = cachefly_log_target.s3_logs.name
      type      = cachefly_log_target.s3_logs.type
      created   = cachefly_log_target.s3_logs.created_at
      updated   = cachefly_log_target.s3_logs.updated_at
      services  = {
        access = cachefly_log_target.s3_logs.access_logs_services
        origin = cachefly_log_target.s3_logs.origin_logs_services
      }
    }
    elasticsearch = {
      id      = cachefly_log_target.elasticsearch_logs.id
      name    = cachefly_log_target.elasticsearch_logs.name
      type    = cachefly_log_target.elasticsearch_logs.type
      hosts   = cachefly_log_target.elasticsearch_logs.hosts
      index   = cachefly_log_target.elasticsearch_logs.index
      created = cachefly_log_target.elasticsearch_logs.created_at
      updated = cachefly_log_target.elasticsearch_logs.updated_at
    }
  }
}
