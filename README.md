<p align="center">
 <img src="https://www.cachefly.com/wp-content/uploads/2023/10/Thumbnail-About-Us-Video.png" alt="CacheFly Logo" width="200"/>
</p>

<h4 align="center">Terraform Provider for CacheFly API (2.5.0)</h4>

<hr style="width: 50%; border: 1px solid #000; margin: 20px auto;">

# Terraform Provider for CacheFly

A Golang Terraform provider built using CacheFly Go SDK [GoLang SDK for CacheFly v1.0](https://github.com/cachefly/cachfly-go-sdk). Supports CacheFly API (2.5.0)

> CacheFly API v2.6 support coming soon.

## About CacheFly

CacheFly CDN is the only CDN built for throughput, delivering rich-media content up to 158% faster than other major CDNs.

## Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    cachefly = {
      source  = "cachefly/cachefly" 
      version = "0.1.0"
    }
  }
}

provider "cachefly" {
  api_token = "your-cachefly-api-token" # Replace with a valid CacheFly API token
}
```

## Quick Start

Here's a basic example to get you started:

```hcl
# List existing services
data "cachefly_services" "list_services" {
  response_type     = "shallow"
  include_features  = false
  status           = "ACTIVE"
  offset           = 0
  limit            = 10
}

output "services" {
  value = data.cachefly_services.list_services.services
}
```

## Examples

Explore comprehensive examples to help you get started:

- **[Quick Start Guide](./examples/quickstart-setup/main.tf)** - Simple examples to get up and running quickly
- **[Origins Configuration](./examples/origins-setup/main.tf)** - Setting up origin servers and configurations  
- **[CDN Service + Domain Configuration](./examples/service-domain-setup/main.tf)** - Setting up CDN Service and Domains

- **[SSL Certificate Configuration](./examples/certificate-setup/main.tf)** - Example shows how to configure SSL certificate
- **[Script Configuration](./examples/script-config-setup/main.tf)** - CDN Service advanced script congigrations

- **[Users Setup](./examples/users-setup/main.tf)** - Setting user accounts 
- **[Service Options](./examples/service-options-advanced-setup/main.tf)** - Advanced service configuration options
- **[Service, Domain and Options](./examples/service-domain-options-setup/main.tf)** - Full configuration service, domain and options

- **[Comprehensive Setup](./examples/comprehensive-setup/main.tf)** - Full configuration with service creation and management


## Features
- **Service Management** - Create, update, and manage CacheFly services with auto-SSL
- **Origin Configuration** - Configure origin servers with timeouts, compression, and caching settings
- **Custom Domain Management** - Attach custom domains with DNS validation
- **SSL Certificate Management** - Custom SSL certificates and auto-SSL configuration
- **Comprehensive Service Options** - 30+ CDN configuration options including:
 * Caching Control (nocache, servestale, normalizequerystring, purgenoquery)
 * Performance Optimization (brotli_support, bandwidth throttling, connection limits)
 * Geographic Caching (by country, region, referer)
 * Timeout Configurations (TTFB, connection, error TTL, max connections)
 * Security Features (API keys, serve key protection)
 * Advanced Features (live streaming, link preheating, CORS, redirects)
 * File Handling (skip encoding/processing by extension, directory purge control)
- **Script Configurations** - Automation and custom logic (URL redirects, AWS credentials)
- **User Management** - Team access control with granular permissions and service assignments
- **Full API Coverage** - Comprehensive support for CacheFly API v2.5.0

## Resources and Data Sources

### Resources
- `cachefly_service` - Manage CacheFly CDN services with auto-SSL and configuration modes
- `cachefly_origin` - Configure origin servers with timeouts, compression, and caching settings
- `cachefly_service_domain` - Attach and manage custom domains with DNS validation
- `cachefly_certificate` - Upload and manage custom SSL certificates
- `cachefly_service_options` - Configure comprehensive CDN options (30+ settings)
- `cachefly_script_config` - Manage automation scripts and custom logic configurations
- `cachefly_user` - Create and manage team users with granular permissions

### Data Sources
- `cachefly_services` - List and query existing services
- `cachefly_service` - Get details of a specific service
- `cachefly_origins` - List and query origin servers by type
- `cachefly_service_domains` - List domains attached to a specific service

## Documentation

For detailed documentation on all resources and data sources, visit the [Terraform Registry documentation](https://registry.terraform.io/providers/cachefly.com/avvvet/cachefly/latest/docs).

## Output Example

The following screenshot shows Terraform in action, deploying a new service and retrieving the newly created service:

![Terraform output for CacheFly services](./hcl_output.png)

## Requirements

- Terraform >= 0.13
- Go >= 1.19 (for development)
- Valid CacheFly API token

## Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests to help improve this provider.

## Support

For issues and questions:
- Open an issue on [GitHub](https://github.com/cachefly/terraform-provider-cachefly/issues)
- Contact CacheFly support for API-related questions

## License

This project is licensed under the MIT License - see the LICENSE file for details.