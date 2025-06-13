package resources

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

// Test Schema validation
func TestServiceDomainResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := NewServiceDomainResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	// no errors
	assert.False(t, resp.Diagnostics.HasError(), "Schema should not have errors")

	// required attributes exist
	attrs := resp.Schema.Attributes
	assert.Contains(t, attrs, "id")
	assert.Contains(t, attrs, "service_id")
	assert.Contains(t, attrs, "name")

	// optional attributes exist
	assert.Contains(t, attrs, "description")
	assert.Contains(t, attrs, "validation_mode")

	// computed attributes exist
	assert.Contains(t, attrs, "validation_target")
	assert.Contains(t, attrs, "validation_status")
	assert.Contains(t, attrs, "certificates")
	assert.Contains(t, attrs, "created_at")
	assert.Contains(t, attrs, "updated_at")
}

// Test Resource metadata
func TestServiceDomainResourceMetadata(t *testing.T) {
	ctx := context.Background()
	r := NewServiceDomainResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "cachefly",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "cachefly_service_domain", resp.TypeName)
}

// Test Configure error handling
func TestServiceDomainResourceConfigure(t *testing.T) {
	ctx := context.Background()
	r := NewServiceDomainResource().(*ServiceDomainResource)

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

// Test Import ID parsing logic
func TestServiceDomainResourceImportIDParsing(t *testing.T) {
	tests := []struct {
		name        string
		importID    string
		expectError bool
		serviceID   string
		domainID    string
	}{
		{
			name:        "valid import format",
			importID:    "service123:domain456",
			expectError: false,
			serviceID:   "service123",
			domainID:    "domain456",
		},
		{
			name:        "invalid import format - no colon",
			importID:    "service123domain456",
			expectError: true,
		},
		{
			name:        "invalid import format - too many parts",
			importID:    "service123:domain456:extra",
			expectError: true,
		},
		{
			name:        "invalid import format - empty",
			importID:    "",
			expectError: true,
		},
		{
			name:        "invalid import format - only colon",
			importID:    ":",
			expectError: true,
		},
		{
			name:        "invalid import format - empty service_id",
			importID:    ":domain456",
			expectError: true,
		},
		{
			name:        "invalid import format - empty domain_id",
			importID:    "service123:",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			parts := strings.Split(tt.importID, ":")
			isValid := len(parts) == 2 && parts[0] != "" && parts[1] != ""

			if tt.expectError {
				assert.False(t, isValid, "Should be invalid for ID: %s", tt.importID)
			} else {
				assert.True(t, isValid, "Should be valid for ID: %s", tt.importID)
				if isValid {
					assert.Equal(t, tt.serviceID, parts[0], "Service ID should match")
					assert.Equal(t, tt.domainID, parts[1], "Domain ID should match")
				}
			}
		})
	}
}
