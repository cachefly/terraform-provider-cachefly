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
      source  = "cachefly.com/avvvet/cachefly"
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

- **[Complete Setup](./examples/complete-setup/main.tf)** - Full configuration with service creation and management
- **[Origins Configuration](./examples/origins/main.tf)** - Setting up origin servers and configurations  
- **[Quick Start Guide](./examples/quickstart/main.tf)** - Simple examples to get up and running quickly
- **[Service Options](./examples/service-options/main.tf)** - Advanced service configuration options

## Features

- **Service Management** - Create, update, and manage CacheFly services
- **Origin Configuration** - Configure origin servers and settings
- **Data Sources** - Query existing services and configurations
- **Full API Coverage** - Comprehensive support for CacheFly API v2.5.0

## Resources and Data Sources

### Resources
- `cachefly_service` - Manage CacheFly services
- `cachefly_origin` - Configure origin servers

### Data Sources  
- `cachefly_services` - List and query existing services
- `cachefly_service` - Get details of a specific service

## Documentation

For detailed documentation on all resources and data sources, visit the [Terraform Registry documentation](https://registry.terraform.io/providers/cachefly.com/avvvet/cachefly/latest/docs).

## Output Example

The following screenshot shows Terraform in action, deploying a new service and retrieving the newly created service:

![Terraform output for CacheFly services](./doc/hcl_output.png)

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