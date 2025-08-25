package models

import (
	"context"

	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ScriptConfigModel represents the Terraform model for a CacheFly script config
type ScriptConfigModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Services               types.Set    `tfsdk:"services"`                 // Set of service IDs
	ScriptConfigDefinition types.String `tfsdk:"script_config_definition"` // Global script config definition ID
	MimeType               types.String `tfsdk:"mime_type"`
	Value                  types.String `tfsdk:"value"`
	Activated              types.Bool   `tfsdk:"activated"`
	Purpose                types.String `tfsdk:"purpose"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
	Status                 types.String `tfsdk:"status"`
	UseSchema              types.Bool   `tfsdk:"use_schema"`
	DataModel              types.String `tfsdk:"data_model"`
}

// ScriptConfigActivationModel represents the Terraform model for script config activation
type ScriptConfigActivationModel struct {
	ID             types.String `tfsdk:"id"`
	ScriptConfigID types.String `tfsdk:"script_config_id"`
	Activated      types.Bool   `tfsdk:"activated"`
}

// ToSDKCreateRequest converts the Terraform model to SDK CreateScriptConfigRequest
func (m *ScriptConfigModel) ToSDKCreateRequest(ctx context.Context) *api.CreateScriptConfigRequest {
	// Convert Services set to string slice
	var services []string
	if !m.Services.IsUnknown() {
		serviceElements := make([]types.String, 0, len(m.Services.Elements()))
		m.Services.ElementsAs(ctx, &serviceElements, false)
		for _, elem := range serviceElements {
			services = append(services, elem.ValueString())
		}
	}

	return &api.CreateScriptConfigRequest{
		Name:                   m.Name.ValueString(),
		Services:               services,
		ScriptConfigDefinition: m.ScriptConfigDefinition.ValueString(),
		MimeType:               m.MimeType.ValueString(),
		Value:                  m.Value.ValueString(),
	}
}
