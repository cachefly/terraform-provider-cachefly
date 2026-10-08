package models

import (
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEdgeControlKVDataToAPI(t *testing.T) {
	tests := []struct {
		name    string
		value   types.Dynamic
		want    map[string]interface{}
		wantErr string
	}{
		{
			name: "object of scalars",
			value: types.DynamicValue(types.ObjectValueMust(
				map[string]attr.Type{"region": types.StringType, "ttl": types.NumberType, "ratio": types.NumberType, "on": types.BoolType},
				map[string]attr.Value{
					"region": types.StringValue("eu"),
					"ttl":    types.NumberValue(big.NewFloat(300)),
					"ratio":  types.NumberValue(big.NewFloat(1.5)),
					"on":     types.BoolValue(true),
				},
			)),
			want: map[string]interface{}{"region": "eu", "ttl": int64(300), "ratio": 1.5, "on": true},
		},
		{
			name:  "map",
			value: types.DynamicValue(types.MapValueMust(types.StringType, map[string]attr.Value{"a": types.StringValue("x")})),
			want:  map[string]interface{}{"a": "x"},
		},
		{
			name: "null value is sent as null",
			value: types.DynamicValue(types.ObjectValueMust(
				map[string]attr.Type{"a": types.StringType},
				map[string]attr.Value{"a": types.StringNull()},
			)),
			want: map[string]interface{}{"a": nil},
		},
		{
			name: "unknown values are skipped",
			value: types.DynamicValue(types.ObjectValueMust(
				map[string]attr.Type{"a": types.StringType, "b": types.StringType},
				map[string]attr.Value{"a": types.StringUnknown(), "b": types.StringValue("y")},
			)),
			want: map[string]interface{}{"b": "y"},
		},
		{
			name: "dynamic element",
			value: types.DynamicValue(types.ObjectValueMust(
				map[string]attr.Type{"a": types.DynamicType},
				map[string]attr.Value{"a": types.DynamicValue(types.BoolValue(false))},
			)),
			want: map[string]interface{}{"a": false},
		},
		{
			name:  "empty object",
			value: types.DynamicValue(types.ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{})),
			want:  map[string]interface{}{},
		},
		{
			name:  "null data",
			value: types.DynamicNull(),
			want:  map[string]interface{}{},
		},
		{
			name: "nested list is rejected",
			value: types.DynamicValue(types.ObjectValueMust(
				map[string]attr.Type{"list": types.ListType{ElemType: types.StringType}},
				map[string]attr.Value{"list": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("x")})},
			)),
			wantErr: `value for key "list" must be a string, number or boolean`,
		},
		{
			name:    "non-object data is rejected",
			value:   types.DynamicValue(types.StringValue("x")),
			wantErr: "data must be an object of keys and values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EdgeControlKVDataToAPI(tt.value)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEdgeControlKVDataFromAPI(t *testing.T) {
	value, err := EdgeControlKVDataFromAPI(map[string]interface{}{
		"region": "eu",
		"ttl":    float64(300),
		"ratio":  0.1,
		"on":     true,
		"none":   nil,
	})
	require.NoError(t, err)

	object, ok := value.UnderlyingValue().(basetypes.ObjectValue)
	require.True(t, ok, "expected an object, got %T", value.UnderlyingValue())
	attributes := object.Attributes()

	assert.Equal(t, types.StringValue("eu"), attributes["region"])
	assert.Equal(t, types.BoolValue(true), attributes["on"])
	assert.Equal(t, types.StringNull(), attributes["none"])

	ttl, ok := attributes["ttl"].(basetypes.NumberValue)
	require.True(t, ok)
	assert.Equal(t, 0, ttl.ValueBigFloat().Cmp(big.NewFloat(300)))

	// 0.1 must equal the number Terraform parses from "0.1" in configuration,
	// not the binary float64 approximation.
	configured, _, err := big.ParseFloat("0.1", 10, 512, big.ToNearestEven)
	require.NoError(t, err)
	ratio, ok := attributes["ratio"].(basetypes.NumberValue)
	require.True(t, ok)
	assert.Equal(t, 0, ratio.ValueBigFloat().Cmp(configured))
}

func TestEdgeControlKVDataRoundTrip(t *testing.T) {
	apiData := map[string]interface{}{"region": "eu", "ttl": float64(300), "ratio": 0.25, "on": false}

	value, err := EdgeControlKVDataFromAPI(apiData)
	require.NoError(t, err)

	got, err := EdgeControlKVDataToAPI(value)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"region": "eu", "ttl": int64(300), "ratio": 0.25, "on": false}, got)
}
