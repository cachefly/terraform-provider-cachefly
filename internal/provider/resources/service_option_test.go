package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

// Test Schema validation
func TestServiceOptionsResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := NewServiceOptionsResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	// no errors
	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	// required attributes exist
	attrs := resp.Schema.Attributes
	assert.Contains(t, attrs, "service_id")

	assert.Contains(t, attrs, "options")
	assert.Contains(t, attrs, "last_updated")

	assert.True(t, attrs["service_id"].IsRequired())

	assert.True(t, attrs["options"].IsOptional())

	assert.True(t, attrs["last_updated"].IsComputed())
}

// Test Resource metadata
func TestServiceOptionsResourceMetadata(t *testing.T) {
	ctx := context.Background()
	r := NewServiceOptionsResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "cachefly",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "cachefly_service_options", resp.TypeName)
}

// Test Configure error handling
func TestServiceOptionsResourceConfigure(t *testing.T) {
	ctx := context.Background()
	r := NewServiceOptionsResource().(*ServiceOptionsResource)

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
