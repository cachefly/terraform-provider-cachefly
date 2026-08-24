package resources

import (
	"testing"

	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

// mapScriptConfigToState must store the value verbatim when the API returns it as
// a raw string (e.g. a YAML document), and must NOT re-encode it as a quoted JSON
// string. Re-encoding double-escapes the value, so the refreshed state differs from
// the configured value on every plan ("inconsistent result after apply" + perpetual
// replace). Regression test for that bug.
func TestMapScriptConfigToState_StringValueNotDoubleEncoded(t *testing.T) {
	r := &ScriptConfigResource{}

	yaml := "\"allow\": []\n\"default\": \"allow\"\n\"deny\":\n- \"meta-externalagent\"\n"

	config := &api.ScriptConfig{
		ID:    "abc123",
		Value: yaml, // API returns the value as a string
	}
	var data models.ScriptConfigModel
	r.mapScriptConfigToState(config, &data)

	if got := data.Value.ValueString(); got != yaml {
		t.Errorf("string value was not stored verbatim.\n want: %q\n  got: %q", yaml, got)
	}
}

// A non-string value (e.g. a JSON object the API parsed into a map) should still be
// marshalled to a JSON string, preserving the pre-fix behaviour for that case.
func TestMapScriptConfigToState_ObjectValueMarshalled(t *testing.T) {
	r := &ScriptConfigResource{}

	config := &api.ScriptConfig{
		ID:    "abc123",
		Value: map[string]interface{}{"something": float64(2324)},
	}
	var data models.ScriptConfigModel
	r.mapScriptConfigToState(config, &data)

	if got, want := data.Value.ValueString(), `{"something":2324}`; got != want {
		t.Errorf("object value not marshalled as expected.\n want: %q\n  got: %q", want, got)
	}
}

// A nil value maps to null.
func TestMapScriptConfigToState_NilValueIsNull(t *testing.T) {
	r := &ScriptConfigResource{}

	config := &api.ScriptConfig{ID: "abc123", Value: nil}
	var data models.ScriptConfigModel
	r.mapScriptConfigToState(config, &data)

	if !data.Value.IsNull() {
		t.Errorf("nil value should map to null, got %q", data.Value.ValueString())
	}
}
