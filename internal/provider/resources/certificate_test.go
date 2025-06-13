package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

// Test 1: Schema validation
func TestCertificateResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := NewCertificateResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	// Verify no errors
	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	attrs := resp.Schema.Attributes
	assert.Contains(t, attrs, "certificate")
	assert.Contains(t, attrs, "certificate_key")

	assert.Contains(t, attrs, "password")

	assert.Contains(t, attrs, "id")
	assert.Contains(t, attrs, "subject_common_name")
	assert.Contains(t, attrs, "subject_names")
	assert.Contains(t, attrs, "expired")
	assert.Contains(t, attrs, "expiring")
	assert.Contains(t, attrs, "in_use")
	assert.Contains(t, attrs, "managed")
	assert.Contains(t, attrs, "services")
	assert.Contains(t, attrs, "domains")
	assert.Contains(t, attrs, "not_before")
	assert.Contains(t, attrs, "not_after")
	assert.Contains(t, attrs, "created_at")

	assert.True(t, attrs["certificate"].IsRequired())
	assert.True(t, attrs["certificate_key"].IsRequired())

	assert.True(t, attrs["password"].IsOptional())

	assert.True(t, attrs["id"].IsComputed())
	assert.True(t, attrs["subject_common_name"].IsComputed())
	assert.True(t, attrs["expired"].IsComputed())

	assert.True(t, attrs["certificate"].IsSensitive())
	assert.True(t, attrs["certificate_key"].IsSensitive())
	assert.True(t, attrs["password"].IsSensitive())
}

// Test 2: Resource metadata
func TestCertificateResourceMetadata(t *testing.T) {
	ctx := context.Background()
	r := NewCertificateResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "cachefly",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "cachefly_certificate", resp.TypeName)
}

// Test 3: Configure error handling
func TestCertificateResourceConfigure(t *testing.T) {
	ctx := context.Background()
	r := NewCertificateResource().(*CertificateResource)

	// Test with nil provider data
	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}
	r.Configure(ctx, req, resp)
	assert.False(t, resp.Diagnostics.HasError(), "Should not error with nil provider data")

	// Test with wrong type should error
	req.ProviderData = "wrong-type"
	resp = &resource.ConfigureResponse{}
	r.Configure(ctx, req, resp)
	assert.True(t, resp.Diagnostics.HasError(), "Should error with wrong provider data type")
}
