package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/cachefly/cachefly-sdk-go/pkg/cachefly"
)

var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cachefly": providerserver.NewProtocol6WithError(New("test")()),
}

func TestAccPreCheck(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	if v := os.Getenv("CACHEFLY_API_TOKEN"); v == "" {
		t.Fatal("CACHEFLY_API_TOKEN must be set for acceptance tests")
	}
}

// TestAccEdgeControlPreCheck runs TestAccPreCheck and skips the test when Edge
// Control is not enabled for the test account.
func TestAccEdgeControlPreCheck(t *testing.T) {
	TestAccPreCheck(t)

	if _, err := GetSDKClient().EdgeControlKV.GetAccount(context.Background()); err != nil {
		if strings.Contains(err.Error(), "Edge Control is not enabled") {
			t.Skip("Edge Control is not enabled for the test account")
		}
		t.Fatalf("Could not check whether Edge Control is enabled: %v", err)
	}
}

type TestOrganizationType string

// Define values for TestOrganizationType
const (
	HOSTED     TestOrganizationType = "HOSTED"
	HYBRID     TestOrganizationType = "HYBRID"
	HOSTEDSCIM TestOrganizationType = "HOSTED_SCIM"
)

func ProviderConfig(t *testing.T, testOrganizationType TestOrganizationType) string {
	var token string
	switch testOrganizationType {
	case HOSTED:
		token = os.Getenv("HOSTED_ORGANIZATION_API_TOKEN")
	case HOSTEDSCIM:
		token = os.Getenv("HOSTED_SCIM_ORGANIZATION_API_TOKEN")
	case HYBRID:
		token = os.Getenv("HYBRID_ORGANIZATION_API_TOKEN")
	default:
		t.Fatalf("Invalid test organization type: %v", testOrganizationType)
	}

	return fmt.Sprintf(`
provider "cachefly" {
	api_token = "%v"
	base_url = "%v"
}`, token, os.Getenv("CACHEFLY_BASE_URL"))
}

var sdkClient *cachefly.Client
var sdkClientOnce sync.Once

func GetSDKClient() *cachefly.Client {
	sdkClientOnce.Do(func() {
		apiToken := os.Getenv("CACHEFLY_API_TOKEN")
		if apiToken == "" {
			panic("CACHEFLY_API_TOKEN environment variable must be set for testing")
		}

		baseURL := os.Getenv("CACHEFLY_BASE_URL")
		if baseURL == "" {
			baseURL = "https://api.cachefly.com/api/2.6"
		}

		sdkClient = cachefly.NewClient(
			cachefly.WithToken(apiToken),
			cachefly.WithBaseURL(baseURL),
		)

		if sdkClient == nil {
			panic("Failed to create CacheFly client for testing")
		}
	})

	return sdkClient
}
