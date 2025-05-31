# CacheFly Origins Example

This example demonstrates how to create and manage origin servers in CacheFly. Origins are backend servers that CacheFly fetches content from.

## What This Creates

### 🌐 **HTTP Origins (3)**
- **Web Server** - Main website backend (`web.example.com`)
- **API Server** - API backend with query param caching (`api.example.com`)
- **Dev Server** - Development server with short cache times (`dev.example.com`)

### ☁️ **S3 Origin (1)**
- **Static Assets** - S3 bucket for images, CSS, JS files

### 📊 **Data Sources**
- Single origin lookup verification
- All origins listing
- Filtered origins by type (S3 only)

## Prerequisites

1. **CacheFly Account** with API access
2. **Backend servers** - The servers you want to cache content from
3. **S3 credentials** (if using S3 origins)

## Quick Setup

### 1. Configure Your Settings

Edit the `main.tf` file:

```hcl
provider "cachefly" {
  api_token = "your-actual-api-token-here"
}

# Update hostnames to your actual servers
resource "cachefly_origin" "web_server" {
  hostname = "your-web-server.com"  # Your actual hostname
  # ... other settings
}
```

### 2. Update S3 Credentials (if using S3)

```hcl
resource "cachefly_origin" "s3_assets" {
  hostname    = "your-bucket.s3.amazonaws.com"
  access_key  = "your-s3-access-key"
  secret_key  = "your-s3-secret-key"
  region      = "us-east-1"
  # ... other settings
}
```

### 3. Apply Configuration

```bash
# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Create the origins
terraform apply
```

## Understanding Origins

### Origin Types

**HTTP Origins** - Standard web servers:
```hcl
resource "cachefly_origin" "web_server" {
  type     = "http"
  hostname = "web.example.com"
  scheme   = "https"  # or "http"
}
```

**S3 Origins** - Amazon S3 buckets:
```hcl
resource "cachefly_origin" "s3_bucket" {
  type         = "s3"
  hostname     = "bucket.s3.amazonaws.com"
  access_key   = "AKIA..."
  secret_key   = "secret..."
  region       = "us-east-1"
}
```

### Cache Settings

**TTL (Time To Live)**:
- `ttl` - How long to cache successful responses
- `missed_ttl` - How long to cache 404/error responses

**Performance**:
- `gzip` - Enable compression
- `cache_by_query_param` - Cache different responses for different query parameters

**Timeouts**:
- `connection_timeout` - Max time to establish connection
- `time_to_first_byte_timeout` - Max time to receive first byte

## Common Configurations

### Web Application Server

```hcl
resource "cachefly_origin" "webapp" {
  type                 = "http"
  hostname             = "app.mycompany.com"
  scheme               = "https"
  cache_by_query_param = false
  gzip                 = true
  ttl                  = 86400  # 24 hours
  missed_ttl           = 300    # 5 minutes
}
```

### API Server

```hcl
resource "cachefly_origin" "api" {
  type                 = "http"
  hostname             = "api.mycompany.com"
  scheme               = "https"
  cache_by_query_param = true   # Cache by query params
  gzip                 = true
  ttl                  = 3600   # 1 hour
  missed_ttl           = 60     # 1 minute
}
```

### Static Assets (S3)

```hcl
resource "cachefly_origin" "assets" {
  type         = "s3"
  hostname     = "assets.s3.amazonaws.com"
  gzip         = true
  ttl          = 259200  # 3 days
  missed_ttl   = 3600    # 1 hour
  
  access_key   = var.s3_access_key
  secret_key   = var.s3_secret_key
  region       = "us-east-1"
}
```

## Verification

After deployment, verify everything is working:

```bash
# Check all outputs
terraform output

# See origins summary
terraform output origins_summary

# Check verification
terraform output verification
```

You should see:
- ✅ `data_source_works = true`
- ✅ `all_origins_created = true`
- ✅ `total_origins >= 4`

## Using Origins with Services

Origins by themselves don't serve traffic - you need to attach them to services:

### Method 1: Via Rules (Future Feature)
```hcl
# This will be available when we implement rules
resource "cachefly_rule" "web_rule" {
  service_id = cachefly_service.web.id
  origin_id  = cachefly_origin.web_server.id
  # ... rule configuration
}
```

### Method 2: Via Service Configuration
Some services can be configured to use specific origins directly.

## Data Source Usage

### Look Up Single Origin

```hcl
data "cachefly_origin" "existing" {
  id = "origin-id-here"
}

output "origin_details" {
  value = {
    name     = data.cachefly_origin.existing.name
    hostname = data.cachefly_origin.existing.hostname
    type     = data.cachefly_origin.existing.type
  }
}
```

### List All Origins

```hcl
data "cachefly_origins" "all" {
  # Optional filters
  type = "http"  # Only HTTP origins
}

output "origin_count" {
  value = length(data.cachefly_origins.all.origins)
}
```

## Import Existing Origins

```bash
# Import existing origin
terraform import cachefly_origin.existing origin-id-here
```

## Troubleshooting

### Connection Issues

1. **Check hostname** - Make sure it's reachable
2. **Verify SSL** - HTTPS origins need valid certificates
3. **Test timeouts** - Adjust if your server is slow

### S3 Issues

1. **Credentials** - Verify access key and secret key
2. **Bucket permissions** - Ensure CacheFly can read your bucket
3. **Region** - Must match your S3 bucket region

### Common Errors

- **"hostname required"** - Hostname field is mandatory
- **"invalid scheme"** - Use "http" or "https"
- **"S3 credentials invalid"** - Check access key/secret key

## Security Best Practices

### S3 Credentials

Use Terraform variables for sensitive data:

```hc