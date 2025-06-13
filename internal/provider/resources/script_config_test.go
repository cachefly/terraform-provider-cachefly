package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

// Schema validation
func TestScriptConfigResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := NewScriptConfigResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	// Verify no errors
	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	// Verify required attributes exist
	attrs := resp.Schema.Attributes
	assert.Contains(t, attrs, "name")
	assert.Contains(t, attrs, "services")
	assert.Contains(t, attrs, "script_config_definition")

	// optional attributes exist
	assert.Contains(t, attrs, "mime_type")
	assert.Contains(t, attrs, "value")
	assert.Contains(t, attrs, "activated")

	// computed attributes exist
	assert.Contains(t, attrs, "id")
	assert.Contains(t, attrs, "purpose")
	assert.Contains(t, attrs, "created_at")
	assert.Contains(t, attrs, "updated_at")

	// required attributes are marked as required
	assert.True(t, attrs["name"].IsRequired())
	assert.True(t, attrs["services"].IsRequired())
	assert.True(t, attrs["script_config_definition"].IsRequired())

	// optional attributes are marked as optional
	assert.True(t, attrs["mime_type"].IsOptional())
	assert.True(t, attrs["value"].IsOptional())
	assert.True(t, attrs["activated"].IsOptional())

	// computed attributes are marked as computed
	assert.True(t, attrs["id"].IsComputed())
	assert.True(t, attrs["purpose"].IsComputed())
	assert.True(t, attrs["created_at"].IsComputed())
	assert.True(t, attrs["updated_at"].IsComputed())

	// attributes with defaults are also computed
	assert.True(t, attrs["mime_type"].IsComputed())
	assert.True(t, attrs["activated"].IsComputed())
}

// Resource metadata
func TestScriptConfigResourceMetadata(t *testing.T) {
	ctx := context.Background()
	r := NewScriptConfigResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "cachefly",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "cachefly_script_config", resp.TypeName)
}

// Configure error handling
func TestScriptConfigResourceConfigure(t *testing.T) {
	ctx := context.Background()
	r := NewScriptConfigResource().(*ScriptConfigResource)

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
