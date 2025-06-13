package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

// Schema validation (covering all script will be extesive)
func TestOriginResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := NewOriginResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	//no errors
	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	// Verify required attributes exist
	attrs := resp.Schema.Attributes
	assert.Contains(t, attrs, "id")
	assert.Contains(t, attrs, "type")
	assert.Contains(t, attrs, "host")

	// optional attributes exist
	assert.Contains(t, attrs, "name")
	assert.Contains(t, attrs, "scheme")
	assert.Contains(t, attrs, "cache_by_query_param")
	assert.Contains(t, attrs, "gzip")
	assert.Contains(t, attrs, "ttl")
	assert.Contains(t, attrs, "missed_ttl")

	// S3-specific attributes exist
	assert.Contains(t, attrs, "access_key")
	assert.Contains(t, attrs, "secret_key")
	assert.Contains(t, attrs, "region")
	assert.Contains(t, attrs, "signature_version")

	//computed attributes exist
	assert.Contains(t, attrs, "created_at")
	assert.Contains(t, attrs, "updated_at")

	// Verify sensitive attributes are marked as sensitive
	assert.True(t, attrs["access_key"].IsSensitive(), "access_key should be marked as sensitive")
	assert.True(t, attrs["secret_key"].IsSensitive(), "secret_key should be marked as sensitive")
}

// Resource metadata
func TestOriginResourceMetadata(t *testing.T) {
	ctx := context.Background()
	r := NewOriginResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "cachefly",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "cachefly_origin", resp.TypeName)
}

// Configure error handling
func TestOriginResourceConfigure(t *testing.T) {
	ctx := context.Background()
	r := NewOriginResource().(*OriginResource)

	// nil provider data (should not error)
	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(ctx, req, resp)
	assert.False(t, resp.Diagnostics.HasError(), "Should not error with nil provider data")

	// wrong type should error
	req.ProviderData = "wrong-type"
	resp = &resource.ConfigureResponse{}

	r.Configure(ctx, req, resp)
	assert.True(t, resp.Diagnostics.HasError(), "Should error with wrong provider data type")
}
