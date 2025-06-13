package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

// Test Schema validation
func TestServiceResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := NewServiceResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	// Verify no errors
	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	// required attributes exist
	attrs := resp.Schema.Attributes
	assert.Contains(t, attrs, "id")
	assert.Contains(t, attrs, "name")
	assert.Contains(t, attrs, "unique_name")

	// optional attributes exist
	assert.Contains(t, attrs, "description")
	assert.Contains(t, attrs, "auto_ssl")
	assert.Contains(t, attrs, "configuration_mode")

	// computed attributes exist
	assert.Contains(t, attrs, "status")
	assert.Contains(t, attrs, "created_at")
	assert.Contains(t, attrs, "updated_at")
}

// Test Resource metadata
func TestServiceResourceMetadata(t *testing.T) {
	ctx := context.Background()
	r := NewServiceResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "cachefly",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "cachefly_service", resp.TypeName)
}

// Test Configure error handling
func TestServiceResourceConfigure(t *testing.T) {
	ctx := context.Background()
	r := NewServiceResource().(*ServiceResource)

	// Test with nil provider data (should not error)
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
