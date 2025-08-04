package models

import (
	"fmt"

	api "github.com/cachefly/cachefly-go-sdk/pkg/cachefly/api/v2_5"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ServiceOptionsModel represents the Terraform resource model for service options
type ServiceOptionsModel struct {
	ServiceID types.String  `tfsdk:"service_id"`
	Options   types.Dynamic `tfsdk:"options"`
}

// OptionPropertyModel represents detailed metadata about an option property for Terraform
type OptionPropertyModel struct {
	Label      types.String  `tfsdk:"label"`
	ID         types.String  `tfsdk:"id"`
	Name       types.String  `tfsdk:"name"`
	Type       types.String  `tfsdk:"type"` // "boolean", "integer", "enum", "bitfield", "strings"
	MaxValue   types.Int64   `tfsdk:"max_value"`
	MinValue   types.Int64   `tfsdk:"min_value"`
	Default    types.Dynamic `tfsdk:"default"`
	EnumValues types.List    `tfsdk:"enum_values"` // []EnumValueModel
	BitFields  types.List    `tfsdk:"bit_fields"`  // []BitFieldModel
	UpdatedAt  types.String  `tfsdk:"updated_at"`
	CreatedAt  types.String  `tfsdk:"created_at"`
}

// EnumValueModel represents possible values for enum type options
type EnumValueModel struct {
	Value types.String `tfsdk:"value"`
	Label types.String `tfsdk:"label"`
}

// BitFieldModel represents bitfield options (like HTTP methods)
type BitFieldModel struct {
	BitPosition types.Int64  `tfsdk:"bit_position"`
	Key         types.String `tfsdk:"key"`
	Label       types.String `tfsdk:"label"`
}

// PromoInfoModel contains promotional/UI information
type PromoInfoModel struct {
	Enabled     types.Bool   `tfsdk:"enabled"`
	Description types.String `tfsdk:"description"`
	Order       types.Int64  `tfsdk:"order"`
}

// OptionMetadataModel describes a complete service option with all its metadata
type OptionMetadataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Template    types.String `tfsdk:"template"`
	Group       types.String `tfsdk:"group"`
	Scope       types.String `tfsdk:"scope"`
	ReadOnly    types.Bool   `tfsdk:"read_only"`
	Type        types.String `tfsdk:"type"`     // "standard", "dynamic"
	Property    types.Object `tfsdk:"property"` // OptionPropertyModel
	Promo       types.Object `tfsdk:"promo"`    // PromoInfoModel
	UpdatedAt   types.String `tfsdk:"updated_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

// ValidationErrorModel represents a validation error for a specific field
type ValidationErrorModel struct {
	Field   types.String `tfsdk:"field"`
	Message types.String `tfsdk:"message"`
	Code    types.String `tfsdk:"code"`
}

// LegacyAPIKeyModel represents API key payload
type LegacyAPIKeyModel struct {
	ServiceID types.String `tfsdk:"service_id"`
	APIKey    types.String `tfsdk:"api_key"`
}

// ProtectServeKeyModel for protectserve
type ProtectServeKeyModel struct {
	ServiceID         types.String `tfsdk:"service_id"`
	ProtectServeKey   types.String `tfsdk:"protect_serve_key"`
	ForceProtectServe types.String `tfsdk:"force_protect_serve"`
	HideSecrets       types.Bool   `tfsdk:"hide_secrets"`
}

// FTPSettingsModel represents FTP settings
type FTPSettingsModel struct {
	ServiceID   types.String `tfsdk:"service_id"`
	FTPPassword types.String `tfsdk:"ftp_password"`
	HideSecrets types.Bool   `tfsdk:"hide_secrets"`
}

// Conversion methods from API types to Terraform models

// FromAPIServiceOptions converts API ServiceOptions to Terraform model
func (m *ServiceOptionsModel) FromAPIServiceOptions(serviceID string, options api.ServiceOptions) error {
	m.ServiceID = types.StringValue(serviceID)

	// Convert map[string]interface{} to types.Dynamic
	if len(options) > 0 {
		// Create a map[string]attr.Value for the dynamic type
		elements := make(map[string]attr.Value)
		attrTypes := make(map[string]attr.Type)

		for key, value := range options {
			// Convert interface{} to appropriate types.Value based on type
			convertedValue, attrType := convertInterfaceToAttrValue(value)
			elements[key] = convertedValue
			attrTypes[key] = attrType
		}

		// Create the object value
		objValue, diags := types.ObjectValue(attrTypes, elements)
		if diags.HasError() {
			return fmt.Errorf("failed to convert options to object: %v", diags.Errors())
		}

		m.Options = types.DynamicValue(objValue)
	} else {
		m.Options = types.DynamicNull()
	}

	return nil
}

// Helper function to convert interface{} to attr.Value and attr.Type
func convertInterfaceToAttrValue(value interface{}) (attr.Value, attr.Type) {
	switch v := value.(type) {
	case string:
		return types.StringValue(v), types.StringType
	case bool:
		return types.BoolValue(v), types.BoolType
	case int:
		return types.Int64Value(int64(v)), types.Int64Type
	case int64:
		return types.Int64Value(v), types.Int64Type
	case float64:
		return types.Float64Value(v), types.Float64Type
	case map[string]interface{}:
		nestedElements := make(map[string]attr.Value)
		nestedAttrTypes := make(map[string]attr.Type)

		for nestedKey, nestedValue := range v {
			nestedAttrValue, nestedAttrType := convertInterfaceToAttrValue(nestedValue)
			nestedElements[nestedKey] = nestedAttrValue
			nestedAttrTypes[nestedKey] = nestedAttrType
		}

		objValue, _ := types.ObjectValue(nestedAttrTypes, nestedElements)
		return objValue, types.ObjectType{AttrTypes: nestedAttrTypes}
	case []interface{}:
		if len(v) == 0 {
			listValue, _ := types.ListValue(types.StringType, []attr.Value{})
			return listValue, types.ListType{ElemType: types.StringType}
		}

		listElements := make([]attr.Value, len(v))
		var elemType attr.Type = types.StringType

		for i, item := range v {
			itemValue, itemType := convertInterfaceToAttrValue(item)
			listElements[i] = itemValue
			if i == 0 {
				elemType = itemType
			}
		}

		listValue, _ := types.ListValue(elemType, listElements)
		return listValue, types.ListType{ElemType: elemType}
	default:
		return types.StringValue(fmt.Sprintf("%v", v)), types.StringType
	}
}

// // ToAPIServiceOptions converts Terraform model to API ServiceOptions
// func (m *ServiceOptionsModel) ToAPIServiceOptions() (api.ServiceOptions, error) {
// 	if m.Options.IsNull() || m.Options.IsUnknown() {
// 		return api.ServiceOptions{}, nil
// 	}

// 	options := make(api.ServiceOptions)

// 	// Extract the underlying value from the dynamic type
// 	underlyingValue := m.Options.UnderlyingValue()

// 	// The underlying value should be an object
// 	if objValue, ok := underlyingValue.(basetypes.ObjectValue); ok {
// 		attributes := objValue.Attributes()

// 		for key, value := range attributes {
// 			// Convert terraform types back to interface{}
// 			convertedValue, err := convertTerraformValueToInterface(value)
// 			if err != nil {
// 				return nil, fmt.Errorf("failed to convert value for key %s: %w", key, err)
// 			}
// 			if convertedValue != nil {
// 				options[key] = convertedValue
// 			}
// 		}
// 	} else {
// 		return nil, fmt.Errorf("expected object value, got %T", underlyingValue)
// 	}

// 	return options, nil
// }

// // Helper function to convert Terraform values to interface{}
// func convertTerraformValueToInterface(value attr.Value) (interface{}, error) {
// 	if value.IsNull() || value.IsUnknown() {
// 		return nil, nil
// 	}

// 	switch v := value.(type) {
// 	case types.Dynamic:
// 		return convertTerraformValueToInterface(v.UnderlyingValue())
// 	case basetypes.StringValue:
// 		return v.ValueString(), nil
// 	case basetypes.BoolValue:
// 		return v.ValueBool(), nil
// 	case basetypes.Int64Value:
// 		return int(v.ValueInt64()), nil
// 	case basetypes.Float64Value:
// 		return v.ValueFloat64(), nil
// 	case basetypes.NumberValue:
// 		bigFloat := v.ValueBigFloat()
// 		if bigFloat.IsInt() {
// 			if val, accuracy := bigFloat.Int64(); accuracy == big.Exact {
// 				return int(val), nil
// 			}
// 		}
// 		if val, accuracy := bigFloat.Float64(); accuracy == big.Exact {
// 			return val, nil
// 		}
// 	case basetypes.ListValue:
// 		return convertElements(v.Elements())
// 	case basetypes.TupleValue:
// 		return convertElements(v.Elements())
// 	case basetypes.ObjectValue:
// 		result := make(map[string]interface{})
// 		for key, attr := range v.Attributes() {
// 			converted, err := convertTerraformValueToInterface(attr)
// 			if err != nil {
// 				return nil, err
// 			}
// 			if converted != nil {
// 				result[key] = converted
// 			}
// 		}
// 		return result, nil
// 	}

// 	// Fallback to string representation
// 	return fmt.Sprintf("%v", value), nil
// }

// func convertElements(elements []attr.Value) ([]interface{}, error) {
// 	result := make([]interface{}, len(elements))
// 	for i, elem := range elements {
// 		converted, err := convertTerraformValueToInterface(elem)
// 		if err != nil {
// 			return nil, err
// 		}
// 		result[i] = converted
// 	}
// 	return result, nil
// }

// FromAPIProtectServeKey converts API response to Terraform model
func (m *ProtectServeKeyModel) FromAPIProtectServeKey(serviceID string, response *api.ProtectServeKeyResponse) {
	m.ServiceID = types.StringValue(serviceID)
	if response != nil {
		m.ProtectServeKey = types.StringValue(response.ProtectServeKey)
		m.ForceProtectServe = types.StringValue(response.ForceProtectServe)
	} else {
		m.ProtectServeKey = types.StringNull()
		m.ForceProtectServe = types.StringNull()
	}
}

// ToAPIUpdateProtectServeRequest converts Terraform model to API request
func (m *ProtectServeKeyModel) ToAPIUpdateProtectServeRequest() api.UpdateProtectServeRequest {
	return api.UpdateProtectServeRequest{
		ForceProtectServe: m.ForceProtectServe.ValueString(),
		ProtectServeKey:   m.ProtectServeKey.ValueString(),
	}
}

// FromAPIFTPSettings converts API response to Terraform model
func (m *FTPSettingsModel) FromAPIFTPSettings(serviceID string, response *api.FTPSettingsResponse) {
	m.ServiceID = types.StringValue(serviceID)
	if response != nil {
		m.FTPPassword = types.StringValue(response.FTPPassword)
	} else {
		m.FTPPassword = types.StringNull()
	}
}

// FromAPIOptionMetadata converts API OptionMetadata to Terraform model
func (m *OptionMetadataModel) FromAPIOptionMetadata(apiMeta *api.OptionMetadata) error {
	if apiMeta == nil {
		return fmt.Errorf("API metadata is nil")
	}

	m.ID = types.StringValue(apiMeta.ID)
	m.Name = types.StringValue(apiMeta.Name)
	m.Title = types.StringValue(apiMeta.Title)
	m.Description = types.StringValue(apiMeta.Description)
	m.Template = types.StringValue(apiMeta.Template)
	m.Group = types.StringValue(apiMeta.Group)
	m.Scope = types.StringValue(apiMeta.Scope)
	m.ReadOnly = types.BoolValue(apiMeta.ReadOnly)
	m.Type = types.StringValue(apiMeta.Type)
	m.UpdatedAt = types.StringValue(apiMeta.UpdatedAt)
	m.CreatedAt = types.StringValue(apiMeta.CreatedAt)

	// Convert Property if it exists
	if apiMeta.Property != nil {
		propertyModel := OptionPropertyModel{}
		if err := propertyModel.FromAPIOptionProperty(apiMeta.Property); err != nil {
			return fmt.Errorf("failed to convert property: %w", err)
		}

		// Convert to types.Object
		propertyAttrs := map[string]attr.Value{
			"label":       propertyModel.Label,
			"id":          propertyModel.ID,
			"name":        propertyModel.Name,
			"type":        propertyModel.Type,
			"max_value":   propertyModel.MaxValue,
			"min_value":   propertyModel.MinValue,
			"default":     propertyModel.Default,
			"enum_values": propertyModel.EnumValues,
			"bit_fields":  propertyModel.BitFields,
			"updated_at":  propertyModel.UpdatedAt,
			"created_at":  propertyModel.CreatedAt,
		}

		propertyObj, diags := types.ObjectValue(getOptionPropertyAttrTypes(), propertyAttrs)
		if diags.HasError() {
			return fmt.Errorf("failed to create property object: %v", diags.Errors())
		}
		m.Property = propertyObj
	} else {
		m.Property = types.ObjectNull(getOptionPropertyAttrTypes())
	}

	// Convert Promo
	promoAttrs := map[string]attr.Value{
		"enabled":     types.BoolValue(apiMeta.Promo.Enabled),
		"description": types.StringValue(apiMeta.Promo.Description),
		"order":       types.Int64Value(int64(apiMeta.Promo.Order)),
	}

	promoObj, diags := types.ObjectValue(getPromoInfoAttrTypes(), promoAttrs)
	if diags.HasError() {
		return fmt.Errorf("failed to create promo object: %v", diags.Errors())
	}
	m.Promo = promoObj

	return nil
}

// FromAPIOptionProperty converts API OptionProperty to Terraform model
func (m *OptionPropertyModel) FromAPIOptionProperty(apiProp *api.OptionProperty) error {
	if apiProp == nil {
		return fmt.Errorf("API property is nil")
	}

	m.Label = types.StringValue(apiProp.Label)
	m.ID = types.StringValue(apiProp.ID)
	m.Name = types.StringValue(apiProp.Name)
	m.Type = types.StringValue(apiProp.Type)
	m.UpdatedAt = types.StringValue(apiProp.UpdatedAt)
	m.CreatedAt = types.StringValue(apiProp.CreatedAt)

	// Handle optional values
	if apiProp.MaxValue != nil {
		m.MaxValue = types.Int64Value(int64(*apiProp.MaxValue))
	} else {
		m.MaxValue = types.Int64Null()
	}

	if apiProp.MinValue != nil {
		m.MinValue = types.Int64Value(int64(*apiProp.MinValue))
	} else {
		m.MinValue = types.Int64Null()
	}

	// Handle default value
	if apiProp.Default != nil {
		m.Default = types.DynamicValue(basetypes.NewStringValue(fmt.Sprintf("%v", apiProp.Default)))
	} else {
		m.Default = types.DynamicNull()
	}

	// Convert EnumValues
	if len(apiProp.EnumValues) > 0 {
		enumElements := make([]attr.Value, len(apiProp.EnumValues))
		for i, enumVal := range apiProp.EnumValues {
			enumAttrs := map[string]attr.Value{
				"value": types.StringValue(enumVal.Value),
				"label": types.StringValue(enumVal.Label),
			}
			enumObj, diags := types.ObjectValue(getEnumValueAttrTypes(), enumAttrs)
			if diags.HasError() {
				return fmt.Errorf("failed to create enum value object: %v", diags.Errors())
			}
			enumElements[i] = enumObj
		}

		enumList, diags := types.ListValue(types.ObjectType{AttrTypes: getEnumValueAttrTypes()}, enumElements)
		if diags.HasError() {
			return fmt.Errorf("failed to create enum values list: %v", diags.Errors())
		}
		m.EnumValues = enumList
	} else {
		m.EnumValues = types.ListNull(types.ObjectType{AttrTypes: getEnumValueAttrTypes()})
	}

	// Convert BitFields
	if len(apiProp.BitFields) > 0 {
		bitElements := make([]attr.Value, len(apiProp.BitFields))
		for i, bitField := range apiProp.BitFields {
			bitAttrs := map[string]attr.Value{
				"bit_position": types.Int64Value(int64(bitField.BitPosition)),
				"key":          types.StringValue(bitField.Key),
				"label":        types.StringValue(bitField.Label),
			}
			bitObj, diags := types.ObjectValue(getBitFieldAttrTypes(), bitAttrs)
			if diags.HasError() {
				return fmt.Errorf("failed to create bit field object: %v", diags.Errors())
			}
			bitElements[i] = bitObj
		}

		bitList, diags := types.ListValue(types.ObjectType{AttrTypes: getBitFieldAttrTypes()}, bitElements)
		if diags.HasError() {
			return fmt.Errorf("failed to create bit fields list: %v", diags.Errors())
		}
		m.BitFields = bitList
	} else {
		m.BitFields = types.ListNull(types.ObjectType{AttrTypes: getBitFieldAttrTypes()})
	}

	return nil
}

// Helper functions to get attribute types for nested objects

func getOptionPropertyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"label":       types.StringType,
		"id":          types.StringType,
		"name":        types.StringType,
		"type":        types.StringType,
		"max_value":   types.Int64Type,
		"min_value":   types.Int64Type,
		"default":     types.DynamicType,
		"enum_values": types.ListType{ElemType: types.ObjectType{AttrTypes: getEnumValueAttrTypes()}},
		"bit_fields":  types.ListType{ElemType: types.ObjectType{AttrTypes: getBitFieldAttrTypes()}},
		"updated_at":  types.StringType,
		"created_at":  types.StringType,
	}
}

func getPromoInfoAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":     types.BoolType,
		"description": types.StringType,
		"order":       types.Int64Type,
	}
}

func getEnumValueAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"value": types.StringType,
		"label": types.StringType,
	}
}

func getBitFieldAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"bit_position": types.Int64Type,
		"key":          types.StringType,
		"label":        types.StringType,
	}
}
