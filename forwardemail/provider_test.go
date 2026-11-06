package forwardemail

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/forwardemail/forwardemail-api-go/forwardemail"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	tfprotov6 "github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories attaches the framework provider to
// terraform-plugin-testing over protocol version 6.
func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"forwardemail": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	// Require TF_ACC for acceptance tests to run.
	if v := testAccGetEnv("TF_ACC"); v == "" || v == "0" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	// Ensure the API key is provided for acceptance tests.
	apiKey := testAccGetEnv("FORWARDEMAIL_API_KEY")
	if apiKey == "" {
		t.Skip("FORWARDEMAIL_API_KEY must be set for acceptance tests")
	}
	// Probe account to detect plan restrictions early and skip gracefully.
	client, err := forwardemail.NewClient(apiKey)
	if err != nil {
		t.Fatalf("precheck failed creating forward email client: %v", err)
	}
	if _, err := client.GetAccount(context.Background()); err != nil {
		if strings.Contains(err.Error(), "status: 402") || strings.Contains(err.Error(), "Payment Required") {
			t.Skip("Acceptance tests skipped: Forward Email plan does not permit this operation (402 Payment Required)")
		}
		t.Fatalf("precheck failed getting account: %v", err)
	}
}

func TestAccTests(t *testing.T) {
	for name, c := range TestCases {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			resource.Test(t, c(t))
		})
	}
}

var TestCases = map[string]func(*testing.T) resource.TestCase{
	"account_data": testAccDataSourceForwardemailAccount,
	"domain":       testAccDomainResource,
	"alias":        testAccAliasResource,
}

// testAccGetEnv is a small indirection to allow tests to stub environment access if needed.
func testAccGetEnv(key string) string {
	return os.Getenv(key)
}
