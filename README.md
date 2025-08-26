<p align="center">
 <img src="https://www.cachefly.com/wp-content/uploads/2023/10/Thumbnail-About-Us-Video.png" alt="CacheFly Logo" width="200"/>
</p>

<h4 align="center">Terraform Provider for CacheFly API (2.6.0)</h4>

<hr style="width: 50%; border: 1px solid #000; margin: 20px auto;">

# Terraform Provider for CacheFly

A Golang Terraform provider built using CacheFly Go SDK [GoLang SDK for CacheFly v1.0](https://github.com/cachefly/cachefly-sdk-go). Supports CacheFly API (2.6.0)

## About CacheFly

CacheFly CDN is the only CDN built for throughput, delivering rich-media content up to 158% faster than other major CDNs.

## Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    cachefly = {
      source  = "cachefly/cachefly"
      # Pin to a version or range as needed, e.g.:
      # version = ">= 0.1.0"
    }
  }
}

provider "cachefly" {
  # Prefer environment variable: CACHEFLY_API_TOKEN
  # api_token = "your-cachefly-api-token"

  # Optional: override API base URL (defaults to https://api.cachefly.com/api/2.6)
  # base_url = "https://api.cachefly.com/api/2.6"
}
```

## Quick Start

Basic examples to get started:

```hcl
# Create a CacheFly service
resource "cachefly_service" "example" {
  name        = "Example Service"
  unique_name = "example-service-123"
  auto_ssl    = true

  # Configure service options as a map (see docs for full catalog)
  # options = {
  #   cors = { enabled = true }
  # }
}

output "service_id" {
  value = cachefly_service.example.id
}
```

Or look up an existing service by unique name:

```hcl
data "cachefly_service" "by_unique_name" {
  unique_name      = "example-service-123"
  response_type    = "shallow"
  include_features = false
}

output "service_status" {
  value = data.cachefly_service.by_unique_name.status
}
```

## Examples

Explore examples in this repository:

- **Provider and basics**: [`examples/README.md`](./examples/README.md)
- **Service resource**: [`examples/resources/cachefly_service/resource.tf`](./examples/resources/cachefly_service/resource.tf)
- **Service domain resource**: [`examples/resources/cachefly_service_domain/resource.tf`](./examples/resources/cachefly_service_domain/resource.tf)
- **Origin resource**: [`examples/resources/cachefly_origin/resource.tf`](./examples/resources/cachefly_origin/resource.tf)
- **Certificate resource**: [`examples/resources/cachefly_certificate/resource.tf`](./examples/resources/cachefly_certificate/resource.tf)
- **Script config resource**: [`examples/resources/cachefly_script_config/resource.tf`](./examples/resources/cachefly_script_config/resource.tf)
- **User resource**: [`examples/resources/cachefly_user/resource.tf`](./examples/resources/cachefly_user/resource.tf)
- **Log target resource**: [`examples/resources/cachefly_log_target/resource.tf`](./examples/resources/cachefly_log_target/resource.tf)


## Features
- **Service management**: Create/update CacheFly services with auto-SSL and rich options
- **Origin configuration**: HTTP/S3 origins, timeouts, compression, TTLs
- **Custom domains**: Attach/manage domains with validation and certificates
- **SSL certificates**: Upload and manage custom TLS/SSL certs
- **Script configs**: Manage reusable script configurations and activation
- **Users**: Create/manage users, permissions, and service assignments
- **Log targets**: Configure S3/Elasticsearch/Google Bucket logging targets
- **Data sources**: Query services, domains, origins, users, log targets, delivery regions

## Resources and Data Sources

### Resources
- `cachefly_service`
- `cachefly_service_domain`
- `cachefly_origin`
- `cachefly_user`
- `cachefly_script_config`
- `cachefly_certificate`
- `cachefly_log_target`

### Data Sources
- `cachefly_service`
- `cachefly_service_domain`
- `cachefly_service_domains`
- `cachefly_origin`
- `cachefly_origins`
- `cachefly_log_targets`
- `cachefly_users`
- `cachefly_delivery_regions`

##  Tests
We have unit tests for the provider at `./internal/provider/`.

```bash
go test -v -count=1 ./internal/provider/
```


## Documentation

For detailed documentation on all resources and data sources, visit the [Terraform Registry documentation](https://registry.terraform.io/providers/cachefly/cachefly/latest).

## Requirements

- Terraform >= 0.13
- Go >= 1.23 (for development)
- Valid CacheFly API token (set `CACHEFLY_API_TOKEN` or `api_token`)

## Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests to help improve this provider.

## Support

For issues and questions:
- Open an issue on [GitHub](https://github.com/cachefly/terraform-provider-cachefly/issues)
- Contact CacheFly support for API-related questions

## License

This project is licensed under the MIT License - see the LICENSE file for details.