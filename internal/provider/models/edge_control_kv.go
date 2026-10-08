package models

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// EdgeControlKVResourceModel is the Terraform model for the
// cachefly_edge_control_kv resource.
type EdgeControlKVResourceModel struct {
	ID        types.String  `tfsdk:"id"`
	ServiceID types.String  `tfsdk:"service_id"`
	Data      types.Dynamic `tfsdk:"data"`
	CreatedAt types.String  `tfsdk:"created_at"`
	UpdatedAt types.String  `tfsdk:"updated_at"`
}

// EdgeControlKVDataSourceModel is the Terraform model for the
// cachefly_edge_control_kv data source.
type EdgeControlKVDataSourceModel struct {
	ID        types.String  `tfsdk:"id"`
	ServiceID types.String  `tfsdk:"service_id"`
	Merged    types.Bool    `tfsdk:"merged"`
	Data      types.Dynamic `tfsdk:"data"`
	CreatedAt types.String  `tfsdk:"created_at"`
	UpdatedAt types.String  `tfsdk:"updated_at"`
}

// EdgeControlKVDataToAPI converts KV data to the flat map sent to the API. The
// value must be an object or map whose values are strings, numbers, booleans
// or null. Unknown values are skipped so that configurations can be validated
// before apply.
func EdgeControlKVDataToAPI(value types.Dynamic) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	if value.IsNull() || value.IsUnknown() || value.IsUnderlyingValueNull() || value.IsUnderlyingValueUnknown() {
		return result, nil
	}

	var elements map[string]attr.Value
	switch v := value.UnderlyingValue().(type) {
	case basetypes.ObjectValue:
		elements = v.Attributes()
	case basetypes.MapValue:
		elements = v.Elements()
	default:
		return nil, fmt.Errorf("data must be an object of keys and values, got %s", v.Type(context.Background()))
	}

	for key, element := range elements {
		converted, known, ok := edgeControlKVValueToAPI(element)
		if !ok {
			return nil, fmt.Errorf("value for key %q must be a string, number or boolean, got %s", key, element.Type(context.Background()))
		}
		if known {
			result[key] = converted
		}
	}
	return result, nil
}

// edgeControlKVValueToAPI converts a single KV value. known is false for
// unknown values; ok is false for values that are not scalars.
func edgeControlKVValueToAPI(value attr.Value) (converted interface{}, known bool, ok bool) {
	if value.IsUnknown() {
		return nil, false, true
	}
	if value.IsNull() {
		return nil, true, true
	}

	switch v := value.(type) {
	case basetypes.DynamicValue:
		if v.IsUnderlyingValueUnknown() {
			return nil, false, true
		}
		if v.IsUnderlyingValueNull() {
			return nil, true, true
		}
		return edgeControlKVValueToAPI(v.UnderlyingValue())
	case basetypes.StringValue:
		return v.ValueString(), true, true
	case basetypes.BoolValue:
		return v.ValueBool(), true, true
	case basetypes.NumberValue:
		number := v.ValueBigFloat()
		if number.IsInt() {
			if i, accuracy := number.Int64(); accuracy == big.Exact {
				return i, true, true
			}
		}
		f, _ := number.Float64()
		return f, true, true
	case basetypes.Int64Value:
		return v.ValueInt64(), true, true
	case basetypes.Int32Value:
		return int64(v.ValueInt32()), true, true
	case basetypes.Float64Value:
		return v.ValueFloat64(), true, true
	case basetypes.Float32Value:
		return float64(v.ValueFloat32()), true, true
	}
	return nil, false, false
}

// EdgeControlKVDataFromAPI converts KV data returned by the API to an object
// value. Numbers are parsed from their shortest decimal form so that they equal
// the same numbers written in configuration.
func EdgeControlKVDataFromAPI(data map[string]interface{}) (types.Dynamic, error) {
	attrTypes := make(map[string]attr.Type, len(data))
	attrValues := make(map[string]attr.Value, len(data))

	for key, raw := range data {
		switch v := raw.(type) {
		case nil:
			attrTypes[key] = types.StringType
			attrValues[key] = types.StringNull()
		case string:
			attrTypes[key] = types.StringType
			attrValues[key] = types.StringValue(v)
		case bool:
			attrTypes[key] = types.BoolType
			attrValues[key] = types.BoolValue(v)
		case float64:
			number, err := edgeControlKVNumber(strconv.FormatFloat(v, 'g', -1, 64))
			if err != nil {
				return types.DynamicNull(), fmt.Errorf("invalid number for key %q: %w", key, err)
			}
			attrTypes[key] = types.NumberType
			attrValues[key] = types.NumberValue(number)
		case int:
			attrTypes[key] = types.NumberType
			attrValues[key] = types.NumberValue(new(big.Float).SetInt64(int64(v)))
		case int64:
			attrTypes[key] = types.NumberType
			attrValues[key] = types.NumberValue(new(big.Float).SetInt64(v))
		default:
			return types.DynamicNull(), fmt.Errorf("unsupported value for key %q: %T", key, raw)
		}
	}

	object, diags := types.ObjectValue(attrTypes, attrValues)
	if diags.HasError() {
		return types.DynamicNull(), fmt.Errorf("could not build the data value: %v", diags)
	}
	return types.DynamicValue(object), nil
}

// edgeControlKVNumber parses a decimal number with the precision Terraform uses
// for numbers in configuration.
func edgeControlKVNumber(s string) (*big.Float, error) {
	number, _, err := big.ParseFloat(s, 10, 512, big.ToNearestEven)
	return number, err
}
