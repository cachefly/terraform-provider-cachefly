package resources

import (
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

func TestMapEdgeControlKVToState(t *testing.T) {
	configuredObject := types.DynamicValue(types.ObjectValueMust(
		map[string]attr.Type{"region": types.StringType, "ttl": types.NumberType},
		map[string]attr.Value{"region": types.StringValue("eu"), "ttl": types.NumberValue(big.NewFloat(300))},
	))
	configuredMap := types.DynamicValue(types.MapValueMust(types.StringType, map[string]attr.Value{"region": types.StringValue("eu")}))

	tests := []struct {
		name      string
		serviceID types.String
		current   types.Dynamic
		apiData   map[string]interface{}
		wantID    string
		// wantKept means the current value must be kept as-is.
		wantKept bool
		want     map[string]interface{}
	}{
		{
			name:      "same content keeps the configured object",
			serviceID: types.StringNull(),
			current:   configuredObject,
			apiData:   map[string]interface{}{"ttl": float64(300), "region": "eu"},
			wantID:    "account",
			wantKept:  true,
		},
		{
			name:      "same content keeps a configured map",
			serviceID: types.StringValue("svc-1"),
			current:   configuredMap,
			apiData:   map[string]interface{}{"region": "eu"},
			wantID:    "svc-1",
			wantKept:  true,
		},
		{
			name:      "changed content takes the API data",
			serviceID: types.StringValue("svc-1"),
			current:   configuredObject,
			apiData:   map[string]interface{}{"region": "us", "ttl": float64(300)},
			wantID:    "svc-1",
			want:      map[string]interface{}{"region": "us", "ttl": int64(300)},
		},
		{
			name:      "extra key takes the API data",
			serviceID: types.StringNull(),
			current:   configuredObject,
			apiData:   map[string]interface{}{"region": "eu", "ttl": float64(300), "added": true},
			wantID:    "account",
			want:      map[string]interface{}{"region": "eu", "ttl": int64(300), "added": true},
		},
		{
			name:      "no current value takes the API data",
			serviceID: types.StringValue("svc-1"),
			current:   types.DynamicNull(),
			apiData:   map[string]interface{}{"region": "eu"},
			wantID:    "svc-1",
			want:      map[string]interface{}{"region": "eu"},
		},
		{
			name:      "empty store",
			serviceID: types.StringNull(),
			current:   types.DynamicNull(),
			apiData:   nil,
			wantID:    "account",
			want:      map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := models.EdgeControlKVResourceModel{ServiceID: tt.serviceID, Data: tt.current}
			kv := &api.EdgeControlKV{Data: tt.apiData, CreatedAt: "2026-10-08T10:00:00.000Z", UpdatedAt: "2026-10-08T11:00:00.000Z"}

			require.NoError(t, mapEdgeControlKVToState(kv, &data))

			assert.Equal(t, types.StringValue(tt.wantID), data.ID)
			assert.Equal(t, types.StringValue("2026-10-08T10:00:00.000Z"), data.CreatedAt)
			assert.Equal(t, types.StringValue("2026-10-08T11:00:00.000Z"), data.UpdatedAt)

			if tt.wantKept {
				assert.True(t, data.Data.Equal(tt.current), "expected the current value to be kept, got %s", data.Data)
				return
			}
			_, isObject := data.Data.UnderlyingValue().(basetypes.ObjectValue)
			assert.True(t, isObject, "expected an object value, got %T", data.Data.UnderlyingValue())
			got, err := models.EdgeControlKVDataToAPI(data.Data)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
