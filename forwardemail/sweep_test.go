package forwardemail

import (
	"testing"

	"github.com/forwardemail/terraform-provider-forwardemail/internal/sweep"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestMain registers the resource sweepers and delegates to
// terraform-plugin-testing: when the test binary is invoked with -sweep, only
// the sweepers run; otherwise the normal test suite runs.
func TestMain(m *testing.M) {
	resource.AddTestSweepers("forwardemail_domain", sweep.SweepDomains("forwardemail_domain"))
	resource.TestMain(m)
}
