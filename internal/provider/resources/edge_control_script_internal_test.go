package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"
	api "github.com/cachefly/cachefly-sdk-go/pkg/cachefly/api/v2_6"

	"github.com/cachefly/terraform-provider-cachefly/internal/provider/models"
)

const testEdgeControlScript = "function handler(event) {\n  return event;\n}"

// fakeEdgeControlScripts serves the REQUEST script endpoints of service svc-1
// from memory and records the requests it receives.
type fakeEdgeControlScripts struct {
	t              *testing.T
	mu             sync.Mutex
	versions       []api.EdgeControlScript // versions[i] is version i+1
	requests       []string
	missingService bool
}

func (f *fakeEdgeControlScripts) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	route := strings.TrimPrefix(r.URL.Path, "/edgecontrol/services/svc-1/request")
	f.requests = append(f.requests, r.Method+" "+route)

	if f.missingService {
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
		return
	}

	switch {
	case r.Method == http.MethodGet && route == "/versions/active":
		for _, v := range f.versions {
			if v.Status == api.EdgeControlScriptStatusActive {
				f.write(w, v)
				return
			}
		}
		http.Error(w, `{"message":"No active edge control script for this service."}`, http.StatusNotFound)

	case r.Method == http.MethodGet && route == "/versions":
		newestFirst := make([]api.EdgeControlScript, 0, len(f.versions))
		for i := len(f.versions) - 1; i >= 0; i-- {
			newestFirst = append(newestFirst, f.versions[i])
		}
		f.write(w, newestFirst)

	case r.Method == http.MethodGet && strings.HasPrefix(route, "/versions/"):
		if v := f.version(strings.TrimPrefix(route, "/versions/")); v != nil {
			f.write(w, *v)
			return
		}
		http.Error(w, `{"message":"Version not found."}`, http.StatusNotFound)

	case r.Method == http.MethodPost && route == "/versions":
		var body struct {
			Script string `json:"script"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			f.t.Errorf("Failed to decode request body: %v", err)
		}
		created := api.EdgeControlScript{
			Kind:       string(api.EdgeControlScriptKindRequest),
			Version:    len(f.versions) + 1,
			Script:     body.Script,
			ScriptSize: len(body.Script),
			Status:     api.EdgeControlScriptStatusDeactivated,
		}
		f.versions = append(f.versions, created)
		f.write(w, created)

	case r.Method == http.MethodPut && strings.HasPrefix(route, "/versions/") && strings.HasSuffix(route, "/activate"):
		target := f.version(strings.TrimSuffix(strings.TrimPrefix(route, "/versions/"), "/activate"))
		if target == nil {
			http.Error(w, `{"message":"Version not found for this service."}`, http.StatusBadRequest)
			return
		}
		for i := range f.versions {
			f.versions[i].Status = api.EdgeControlScriptStatusDeactivated
		}
		target.Status = api.EdgeControlScriptStatusActive
		target.LastActivatedAt = "2026-10-08T12:00:00.000Z"
		f.write(w, *target)

	case r.Method == http.MethodPut && route == "/deactivate":
		for i := range f.versions {
			f.versions[i].Status = api.EdgeControlScriptStatusDeactivated
		}
		f.write(w, map[string]string{"message": "Edge control script deactivated."})

	default:
		f.t.Errorf("Unexpected request %s %s", r.Method, r.URL.Path)
		http.Error(w, `{"message":"unexpected"}`, http.StatusInternalServerError)
	}
}

func (f *fakeEdgeControlScripts) version(raw string) *api.EdgeControlScript {
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > len(f.versions) {
		return nil
	}
	return &f.versions[n-1]
}

func (f *fakeEdgeControlScripts) write(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		f.t.Errorf("Failed to encode response: %v", err)
	}
}

// newTestEdgeControlScriptResource returns a resource whose client talks to fake.
func newTestEdgeControlScriptResource(t *testing.T, fake *fakeEdgeControlScripts) *EdgeControlScriptResource {
	t.Helper()
	fake.t = t
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)

	return &EdgeControlScriptResource{
		client: cachefly.NewClient(cachefly.WithToken("test-token"), cachefly.WithBaseURL(server.URL)),
	}
}

func testVersion(n int, status string) api.EdgeControlScript {
	return api.EdgeControlScript{
		Kind:    string(api.EdgeControlScriptKindRequest),
		Version: n,
		Script:  "// version " + strconv.Itoa(n) + "\n" + testEdgeControlScript,
		Status:  status,
	}
}

func TestEdgeControlScriptPublishActivates(t *testing.T) {
	fake := &fakeEdgeControlScripts{}
	r := newTestEdgeControlScriptResource(t, fake)

	script, err := r.publish(context.Background(), "svc-1", api.EdgeControlScriptKindRequest, testEdgeControlScript, true)
	require.NoError(t, err)

	assert.Equal(t, 1, script.Version)
	assert.Equal(t, api.EdgeControlScriptStatusActive, script.Status)
	assert.Equal(t, testEdgeControlScript, script.Script)
	assert.Equal(t, []string{"POST /versions", "PUT /versions/1/activate"}, fake.requests)
}

func TestEdgeControlScriptPublishDeactivated(t *testing.T) {
	fake := &fakeEdgeControlScripts{versions: []api.EdgeControlScript{testVersion(1, api.EdgeControlScriptStatusActive)}}
	r := newTestEdgeControlScriptResource(t, fake)

	script, err := r.publish(context.Background(), "svc-1", api.EdgeControlScriptKindRequest, testEdgeControlScript, false)
	require.NoError(t, err)

	assert.Equal(t, 2, script.Version)
	assert.Equal(t, api.EdgeControlScriptStatusDeactivated, script.Status)
	assert.Equal(t, api.EdgeControlScriptStatusDeactivated, fake.versions[0].Status, "previously active version must be deactivated")
	assert.Equal(t, []string{"POST /versions", "PUT /deactivate", "GET /versions/2"}, fake.requests)
}

func TestEdgeControlScriptSetActivationReactivates(t *testing.T) {
	fake := &fakeEdgeControlScripts{versions: []api.EdgeControlScript{
		testVersion(1, api.EdgeControlScriptStatusDeactivated),
		testVersion(2, api.EdgeControlScriptStatusDeactivated),
	}}
	r := newTestEdgeControlScriptResource(t, fake)

	script, err := r.setActivation(context.Background(), "svc-1", api.EdgeControlScriptKindRequest, 1, true)
	require.NoError(t, err)

	assert.Equal(t, 1, script.Version)
	assert.Equal(t, api.EdgeControlScriptStatusActive, script.Status)
	assert.Equal(t, []string{"PUT /versions/1/activate"}, fake.requests)
}

func TestEdgeControlScriptCurrentScript(t *testing.T) {
	tests := []struct {
		name         string
		versions     []api.EdgeControlScript
		managed      types.Int64
		wantVersion  int
		wantNil      bool
		wantRequests []string
	}{
		{
			name: "active version wins over managed version",
			versions: []api.EdgeControlScript{
				testVersion(1, api.EdgeControlScriptStatusActive),
				testVersion(2, api.EdgeControlScriptStatusDeactivated),
			},
			managed:      types.Int64Value(2),
			wantVersion:  1,
			wantRequests: []string{"GET /versions/active"},
		},
		{
			name: "no active version falls back to managed version",
			versions: []api.EdgeControlScript{
				testVersion(1, api.EdgeControlScriptStatusDeactivated),
				testVersion(2, api.EdgeControlScriptStatusDeactivated),
			},
			managed:      types.Int64Value(1),
			wantVersion:  1,
			wantRequests: []string{"GET /versions/active", "GET /versions/1"},
		},
		{
			name: "unknown managed version falls back to newest version",
			versions: []api.EdgeControlScript{
				testVersion(1, api.EdgeControlScriptStatusDeactivated),
				testVersion(2, api.EdgeControlScriptStatusDeactivated),
			},
			managed:      types.Int64Null(),
			wantVersion:  2,
			wantRequests: []string{"GET /versions/active", "GET /versions"},
		},
		{
			name:         "no versions",
			managed:      types.Int64Null(),
			wantNil:      true,
			wantRequests: []string{"GET /versions/active", "GET /versions"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeEdgeControlScripts{versions: tt.versions}
			r := newTestEdgeControlScriptResource(t, fake)

			script, err := r.currentScript(context.Background(), "svc-1", api.EdgeControlScriptKindRequest, tt.managed)
			require.NoError(t, err)

			if tt.wantNil {
				assert.Nil(t, script)
			} else {
				require.NotNil(t, script)
				assert.Equal(t, tt.wantVersion, script.Version)
			}
			assert.Equal(t, tt.wantRequests, fake.requests)
		})
	}
}

func TestEdgeControlScriptCurrentScriptMissingService(t *testing.T) {
	fake := &fakeEdgeControlScripts{missingService: true}
	r := newTestEdgeControlScriptResource(t, fake)

	script, err := r.currentScript(context.Background(), "svc-1", api.EdgeControlScriptKindRequest, types.Int64Value(3))
	require.NoError(t, err)
	assert.Nil(t, script)
}

func TestMapEdgeControlScriptToState(t *testing.T) {
	var data models.EdgeControlScriptResourceModel
	script := &api.EdgeControlScript{
		Version:         3,
		Script:          testEdgeControlScript,
		ScriptSize:      len(testEdgeControlScript),
		Status:          api.EdgeControlScriptStatusDeactivated,
		Filename:        "abc.js",
		LastActivatedAt: "2026-10-08T12:00:00.000Z",
		CreatedAt:       "2026-10-08T11:00:00.000Z",
	}

	mapEdgeControlScriptToState(script, &data)

	assert.Equal(t, types.Int64Value(3), data.Version)
	assert.Equal(t, types.BoolValue(false), data.Activated)
	assert.Equal(t, types.StringValue(testEdgeControlScript), data.Script)
	assert.Equal(t, types.StringValue("abc.js"), data.Filename)
}
