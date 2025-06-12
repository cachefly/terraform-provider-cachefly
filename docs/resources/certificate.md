---
page_title: "cachefly_certificate Resource - terraform-provider-cachefly"
subcategory: ""
description: |-
  Manages an SSL certificate for CacheFly services.
---

# cachefly_certificate (Resource)

Manages an SSL certificate for CacheFly services.

## Example Usage

```terraform
resource "cachefly_certificate" "main_cert" {
  certificate = <<-EOT
-----BEGIN CERTIFICATE-----

...certificate content here...
-----END CERTIFICATE-----
  EOT
  
  certificate_key = <<-EOT
-----BEGIN PRIVATE KEY-----

...private key content here...
-----END PRIVATE KEY-----
  EOT
}
```

## Schema

### Required

- `certificate` (String, Sensitive) - The SSL certificate in PEM format
- `certificate_key` (String, Sensitive) - The private key for the certificate in PEM format

### Read-Only

- `id` (String) - The unique identifier of the certificate
- `created_at` (String) - When the certificate was created
- `updated_at` (String) - When the certificate was last updated

## Import

Import is supported using the certificate ID:

```shell
terraform import cachefly_certificate.example certificate-id-here
```