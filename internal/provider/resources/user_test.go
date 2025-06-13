package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

// Test Schema validation
func TestUserResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := NewUserResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	attrs := resp.Schema.Attributes
	assert.Contains(t, attrs, "id")
	assert.Contains(t, attrs, "username")
	assert.Contains(t, attrs, "email")
	assert.Contains(t, attrs, "password")
	assert.Contains(t, attrs, "full_name")

	assert.True(t, attrs["password"].IsSensitive())
}

// Test Resource metadata
func TestUserResourceMetadata(t *testing.T) {
	ctx := context.Background()
	r := NewUserResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "cachefly",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "cachefly_user", resp.TypeName)
}

// Test Configure error handling
func TestUserResourceConfigure(t *testing.T) {
	ctx := context.Background()
	r := NewUserResource().(*UserResource) // convert

	// Test with wrong type should error
	req := resource.ConfigureRequest{
		ProviderData: "wrong-type",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(ctx, req, resp)
	assert.True(t, resp.Diagnostics.HasError(), "Should error with wrong provider data type")
}
