# ===================================================================
# CacheFly CDN - example demonstrating  Quickstart CDN service
# 
# ===================================================================

terraform {
  required_version = ">= 1.0"
  required_providers {
    cachefly = {
      source = "cachefly/cachefly" 
      version = "0.0.0"
    }
  }
}


provider "cachefly" {}

resource "cachefly_origin" "test-321321312" {
  name                   = "test-321321312"
  type                   = "WEB"
  hostname               = "example.com"
  scheme                 = "FOLLOW"
}

data "cachefly_origin" "test-321321312" {
  id = cachefly_origin.test-321321312.id
}

