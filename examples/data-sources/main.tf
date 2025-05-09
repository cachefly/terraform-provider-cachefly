terraform {
  required_providers {
    cachefly = {
      source  = "cachefly.com/avvvet/cachefly"
      version = "0.1.0"
    }
  }
}

provider "cachefly" {
  api_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjQ0NDMxLCJ1c2VyIjoiNjgxYjNjZmIyNzE1MzEwMDM1Y2I3MmI2IiwidG9rZW4iOiI2ODFiM2YwNTI3MTUzMTAwMzVjYjdiMjQiLCJpYXQiOjE3NDY2MTYwNjl9.afBXIeoOei7c2pwo7_pbTP1ct-iRJs-8TOaqJAhB2qs" # Replace with a valid CacheFly API token
}

#data "cachefly_services" "example" {}

data "cachefly_services" "list_services" {
  response_type    = "shallow"
  include_features = false
  status           = "ACTIVE"
  offset           = 0
  limit            = 10
}

output "services" {
  value = data.cachefly_services.list_services.services
}