# CacheFly Terraform Provider - Examples

Examples for using the CacheFly Terraform provider to manage your CDN infrastructure.

## 🚀 Quick Start

**New to CacheFly?** Start here: [`quickstart/`](./quickstart/)

This minimal example gets you up and running in under 5 minutes with a basic CDN service.

## 📚 Examples by Development Phase

### Currently Available

- **[quickstart/](./quickstart/)** - Basic CDN service setup (5 minutes)

### In Development

As we develop new resources, you'll find examples here:

- **services/** - Advanced service configurations (available now)
- **origins/** - Backend server setup (coming soon)
- **certificates/** - SSL certificate management (coming soon)
- **domains/** - Custom domain configuration (coming soon)
- **rules/** - Cache and routing rules (coming soon)

### Coming Later

- **complete-setup/** - Full production setup with all resources

## 📋 Prerequisites

1. **CacheFly Account** with API access
2. **API Token** from your CacheFly dashboard  
3. **Terraform** >= 1.0 installed

## 🔧 Setup

### Environment Variable (Recommended)

```bash
export CACHEFLY_API_TOKEN="your-api-token-here"
```

### Provider Configuration

```hcl
terraform {
  required_providers {
    cachefly = {
      source = "cachefly.com/avvvet/cachefly"
    }
  }
}

provider "cachefly" {
  # Uses CACHEFLY_API_TOKEN environment variable
}
```

## 🎯 What's Currently Supported

### ✅ Services Resource

Create and manage CDN services:

```hcl
resource "cachefly_service" "example" {
  name               = "my-service"
  unique_name        = "my-service-01" 
  description        = "My CDN service"
  auto_ssl           = true
  configuration_mode = "API_RULES_AND_OPTIONS"
}
```

### ✅ Services Data Source

Look up existing services:

```hcl
# By ID
data "cachefly_service" "by_id" {
  id = "service-id-here"
}

# By unique name
data "cachefly_service" "by_name" {
  unique_name = "my-service-01"
}
```

## 🔄 In Development

- **Origins** - Backend server configuration
- **Certificates** - SSL certificate management
- **Domains** - Custom domain setup
- **Rules** - Caching and routing rules

## 🛠️ Development Workflow

1. **Start with quickstart** - Get basic functionality working
2. **Use resource-specific examples** - Test individual features during development
3. **Integrate into complete setup** - Combine everything for production use

## 📖 Documentation

- [Provider Documentation](../docs/)
- [CacheFly API Docs](https://docs.cachefly.com/)
- [Terraform Provider Development](https://developer.hashicorp.com/terraform/plugin)

## 🐛 Issues & Support

- **Provider Issues**: [GitHub Issues](https://github.com/avvvet/terraform-provider-cachefly/issues)
- **CacheFly API**: [CacheFly Support](https://cachefly.com/support/)
- **Terraform**: [Terraform Docs](https://terraform.io/docs)

## 🤝 Contributing

Contributing examples:

1. Create examples during resource development
2. Test thoroughly with real CacheFly services
3. Document prerequisites and expected outcomes
4. Include both basic and advanced usage patterns

---

**Ready to get started?** 👉 [Begin with the quickstart example](./quickstart/)