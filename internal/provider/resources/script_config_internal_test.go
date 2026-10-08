package resources

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// decodeJSON decodes s the same way the SDK decodes the API's value field.
func decodeJSON(t *testing.T, s string) interface{} {
	t.Helper()
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("invalid JSON %q: %v", s, err)
	}
	return v
}

func TestScriptConfigValueFromAPI(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  types.String
	}{
		{name: "nil", value: nil, want: types.StringNull()},
		{name: "JSON string is not encoded again", value: `{"301":{"/old":"https://new"}}`, want: types.StringValue(`{"301":{"/old":"https://new"}}`)},
		{name: "YAML string", value: "default:\n  caching: none\n", want: types.StringValue("default:\n  caching: none\n")},
		{name: "object is serialized", value: decodeJSON(t, `{"b": 1, "a": [true, "x"]}`), want: types.StringValue(`{"a":[true,"x"],"b":1}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, scriptConfigValueFromAPI(tt.value))
		})
	}
}

func TestScriptConfigValuesEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{name: "identical text", a: `{"a":1}`, b: `{"a":1}`, want: true},
		{
			name: "whitespace and key order",
			a:    `{"algorithms": [{"name": "CDN77", "type": "PATH", "path": "/data11", "secret": "mysecret1"}]}`,
			b:    `{"algorithms":[{"name":"CDN77","path":"/data11","secret":"mysecret1","type":"PATH"}]}`,
			want: true,
		},
		{name: "escaped ampersand", a: `{"url":"https://x.com/?a=1&b=2"}`, b: `{"url":"https://x.com/?a=1\u0026b=2"}`, want: true},
		{name: "different value", a: `{"a":1}`, b: `{"a":2}`, want: false},
		{name: "array order matters", a: `[1,2]`, b: `[2,1]`, want: false},
		{name: "double-encoded differs from plain", a: `"{\"a\":1}"`, b: `{"a":1}`, want: false},
		{name: "identical YAML", a: "a: 1\n", b: "a: 1\n", want: true},
		{name: "different YAML", a: "a: 1\n", b: "a: 2\n", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, scriptConfigValuesEqual(tt.a, tt.b))
		})
	}
}

func TestMapScriptConfigToStateValue(t *testing.T) {
	const written = `{"algorithms": [{"name": "CDN77", "type": "PATH"}]}`

	tests := []struct {
		name     string
		current  types.String
		apiValue interface{}
		want     types.String
	}{
		{
			name:     "same content as API object keeps current text",
			current:  types.StringValue(written),
			apiValue: decodeJSON(t, `{"algorithms":[{"name":"CDN77","type":"PATH"}]}`),
			want:     types.StringValue(written),
		},
		{
			name:     "same content as API string keeps current text",
			current:  types.StringValue(written),
			apiValue: written,
			want:     types.StringValue(written),
		},
		{
			name:     "double-encoded state is replaced by API string",
			current:  types.StringValue(`"{\"301\":{\"/old\":\"https://new\"}}"`),
			apiValue: `{"301":{"/old":"https://new"}}`,
			want:     types.StringValue(`{"301":{"/old":"https://new"}}`),
		},
		{
			name:     "changed content takes API value",
			current:  types.StringValue(written),
			apiValue: decodeJSON(t, `{"algorithms":[]}`),
			want:     types.StringValue(`{"algorithms":[]}`),
		},
		{
			name:     "no current value takes API value",
			current:  types.StringNull(),
			apiValue: decodeJSON(t, `{"b":1,"a":2}`),
			want:     types.StringValue(`{"a":2,"b":1}`),
		},
		{
			name:     "missing API value clears state",
			current:  types.StringValue(written),
			apiValue: nil,
			want:     types.StringNull(),
		},
	}

	r := &ScriptConfigResource{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := models.ScriptConfigModel{Value: tt.current}
			r.mapScriptConfigToState(&api.ScriptConfig{Value: tt.apiValue}, &data)
			assert.Equal(t, tt.want, data.Value)
		})
	}
}
