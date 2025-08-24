package models

import (
	"fmt"
	"math/big"

	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type ServiceResourceModel struct {
	// Core fields - matches your SDK's Service struct
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	UniqueName  types.String `tfsdk:"unique_name"`
	Description types.String `tfsdk:"description"`

	AutoSSL           types.Bool    `tfsdk:"auto_ssl"`
	ConfigurationMode types.String  `tfsdk:"configuration_mode"`
	TLSProfile        types.String  `tfsdk:"tls_profile"`
	DeliveryRegion    types.String  `tfsdk:"delivery_region"`
	Options           types.Dynamic `tfsdk:"options"`

	//read-only fields
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

type ServiceDataSourceModel struct {
	// Lookup fields (one of these should be provided)
	ID         types.String `tfsdk:"id"`
	UniqueName types.String `tfsdk:"unique_name"`

	// Options for Get() method
	ResponseType    types.String `tfsdk:"response_type"`
	IncludeFeatures types.Bool   `tfsdk:"include_features"`

	Name              types.String  `tfsdk:"name"`
	AutoSSL           types.Bool    `tfsdk:"auto_ssl"`
	ConfigurationMode types.String  `tfsdk:"configuration_mode"`
	Options           types.Dynamic `tfsdk:"options"`
	Status            types.String  `tfsdk:"status"`
	CreatedAt         types.String  `tfsdk:"created_at"`
	UpdatedAt         types.String  `tfsdk:"updated_at"`
}

type ServicesDataSourceModel struct {
	Status          types.String `tfsdk:"status"`
	ResponseType    types.String `tfsdk:"response_type"`
	IncludeFeatures types.Bool   `tfsdk:"include_features"`
	Limit           types.Int64  `tfsdk:"limit"`
	Offset          types.Int64  `tfsdk:"offset"` // maps to ListOptions.Offset

	Services []ServiceListItem `tfsdk:"services"` // maps to ListServicesResponse.Services (data field)
	Meta     ServiceListMeta   `tfsdk:"meta"`
}

type ServiceListItem struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	UniqueName        types.String `tfsdk:"unique_name"`
	AutoSSL           types.Bool   `tfsdk:"auto_ssl"`
	ConfigurationMode types.String `tfsdk:"configuration_mode"`
	Status            types.String `tfsdk:"status"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

type ServiceListMeta struct {
	Limit  types.Int64 `tfsdk:"limit"`  // maps to MetaInfo.Limit
	Offset types.Int64 `tfsdk:"offset"` // maps to MetaInfo.Offset
	Count  types.Int64 `tfsdk:"count"`  // maps to MetaInfo.Count
}

// ToAPIServiceOptions converts Terraform model to API ServiceOptions
func (m *ServiceResourceModel) ToAPIServiceOptions() (api.ServiceOptions, error) {
	if m.Options.IsNull() || m.Options.IsUnknown() {
		return api.ServiceOptions{}, nil
	}

	options := make(api.ServiceOptions)

	underlyingValue := m.Options.UnderlyingValue()

	if objValue, ok := underlyingValue.(basetypes.ObjectValue); ok {
		attributes := objValue.Attributes()

		for key, value := range attributes {
			convertedValue, err := convertTerraformValueToInterface(value)
			if err != nil {
				return nil, fmt.Errorf("failed to convert value for key %s: %w", key, err)
			}
			if convertedValue != nil {
				options[key] = convertedValue
			}
		}
	} else {
		return nil, fmt.Errorf("expected object value, got %T", underlyingValue)
	}

	return options, nil
}

// Helper function to convert Terraform values to interface{}
func convertTerraformValueToInterface(value attr.Value) (interface{}, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	switch v := value.(type) {
	case types.Dynamic:
		return convertTerraformValueToInterface(v.UnderlyingValue())
	case basetypes.StringValue:
		return v.ValueString(), nil
	case basetypes.BoolValue:
		return v.ValueBool(), nil
	case basetypes.Int64Value:
		return int(v.ValueInt64()), nil
	case basetypes.Float64Value:
		return v.ValueFloat64(), nil
	case basetypes.NumberValue:
		bigFloat := v.ValueBigFloat()
		if bigFloat.IsInt() {
			if val, accuracy := bigFloat.Int64(); accuracy == big.Exact {
				return int(val), nil
			}
		}
		if val, accuracy := bigFloat.Float64(); accuracy == big.Exact {
			return val, nil
		}
	case basetypes.ListValue:
		return convertElements(v.Elements())
	case basetypes.TupleValue:
		return convertElements(v.Elements())
	case basetypes.ObjectValue:
		result := make(map[string]interface{})
		for key, attr := range v.Attributes() {
			converted, err := convertTerraformValueToInterface(attr)
			if err != nil {
				return nil, err
			}
			if converted != nil {
				result[key] = converted
			}
		}
		return result, nil
	}

	// Fallback to string representation
	return fmt.Sprintf("%v", value), nil
}

func convertElements(elements []attr.Value) ([]interface{}, error) {
	result := make([]interface{}, len(elements))
	for i, elem := range elements {
		converted, err := convertTerraformValueToInterface(elem)
		if err != nil {
			return nil, err
		}
		result[i] = converted
	}
	return result, nil
}
