# CacheFly Terraform Provider - Examples

Examples for using the CacheFly Terraform provider to manage your CDN infrastructure.

## 🚀 Quick Start

**New to CacheFly?** Start here: [`quickstart/`](./quickstart/)

This minimal example gets you up and running in under 5 minutes with a basic CDN service.

## 📚 Examples by Development Phase

### Currently Available

- **[quickstart/](./quickstart/)** - Basic CDN service setup (5 minutes)


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
      source = "cachefly.com/avvvet/cachefly" # todo: cachefly/cachefly
    }
  }
}

provider "cachefly" {
  # Uses CACHEFLY_API_TOKEN environment variable
}
```

