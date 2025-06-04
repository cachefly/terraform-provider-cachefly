package models

import (
	"context"
	"encoding/json"

	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	if !m.Services.IsNull() && !m.Services.IsUnknown() {
		serviceElements := make([]types.String, 0, len(m.Services.Elements()))
		m.Services.ElementsAs(ctx, &serviceElements, false)
		for _, elem := range serviceElements {
			services = append(services, elem.ValueString())
		}
	}

	// Convert Value JSON string to interface{}
	var value interface{}
	if !m.Value.IsNull() && !m.Value.IsUnknown() && m.Value.ValueString() != "" {
		valueStr := m.Value.ValueString()
		if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
			// If JSON parsing fails, use as string
			value = valueStr
		}
	}

	return &api.CreateScriptConfigRequest{
		Name:                   m.Name.ValueString(),
		Services:               services,
		ScriptConfigDefinition: m.ScriptConfigDefinition.ValueString(),
		MimeType:               m.MimeType.ValueString(),
		Value:                  value,
	}
}

// ToSDKUpdateRequest converts the Terraform model to SDK UpdateScriptConfigRequest
func (m *ScriptConfigModel) ToSDKUpdateRequest(ctx context.Context) *api.UpdateScriptConfigRequest {
	req := &api.UpdateScriptConfigRequest{}

	// Only set fields that have values (not null/unknown)
	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		req.Name = m.Name.ValueString()
	}

	if !m.MimeType.IsNull() && !m.MimeType.IsUnknown() {
		req.MimeType = m.MimeType.ValueString()
	}

	if !m.ScriptConfigDefinition.IsNull() && !m.ScriptConfigDefinition.IsUnknown() {
		req.ScriptConfigDefinition = m.ScriptConfigDefinition.ValueString()
	}

	// Convert Services set to string slice
	if !m.Services.IsNull() && !m.Services.IsUnknown() {
		var services []string
		serviceElements := make([]types.String, 0, len(m.Services.Elements()))
		m.Services.ElementsAs(ctx, &serviceElements, false)
		for _, elem := range serviceElements {
			services = append(services, elem.ValueString())
		}
		req.Services = services
	}

	// Convert Value JSON string to interface{}
	if !m.Value.IsNull() && !m.Value.IsUnknown() {
		valueStr := m.Value.ValueString()
		if valueStr != "" {
			var value interface{}
			if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
				// If JSON parsing fails, use as string
				value = valueStr
			}
			req.Value = value
		}
	}

	return req
}

// FromSDKScriptConfig converts an SDK ScriptConfig to the Terraform model
func (m *ScriptConfigModel) FromSDKScriptConfig(ctx context.Context, config *api.ScriptConfig) {
	m.ID = types.StringValue(config.ID)
	m.Name = types.StringValue(config.Name)
	m.ScriptConfigDefinition = types.StringValue(config.ScriptConfigDefinition)
	m.MimeType = types.StringValue(config.MimeType)
	m.Purpose = types.StringValue(config.Purpose)
	m.CreatedAt = types.StringValue(config.CreatedAt)
	m.UpdatedAt = types.StringValue(config.UpdatedAt)

	// Convert Services slice to set
	if len(config.Services) > 0 {
		serviceValues := make([]attr.Value, len(config.Services))
		for i, service := range config.Services {
			serviceValues[i] = types.StringValue(service)
		}
		m.Services = types.SetValueMust(types.StringType, serviceValues)
	} else {
		m.Services = types.SetValueMust(types.StringType, []attr.Value{})
	}

	// Convert Value interface{} to JSON string
	if config.Value != nil {
		valueBytes, err := json.Marshal(config.Value)
		if err != nil {
			// If marshaling fails, try to convert to string
			if str, ok := config.Value.(string); ok {
				m.Value = types.StringValue(str)
			} else {
				m.Value = types.StringNull()
			}
		} else {
			m.Value = types.StringValue(string(valueBytes))
		}
	} else {
		m.Value = types.StringNull()
	}
}
