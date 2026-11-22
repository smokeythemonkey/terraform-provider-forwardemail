package forwardemail

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/smokeythemonkey/terraform-provider-forwardemail/internal/forwardemail"
)

// createAliasTestServer runs Create against a stub API and captures the
// form parameters the provider sent.
func createAliasTestServer(t *testing.T) (*httptest.Server, *url.Values) {
	t.Helper()
	var gotForm url.Values
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		gotForm = r.PostForm
		fmt.Fprint(w, `{
			"name": "postmaster",
			"domain": {"name": "akbnet.org"},
			"is_enabled": true,
			"recipients": []
		}`)
	}))
	t.Cleanup(svr.Close)
	return svr, &gotForm
}

// runAliasCreate prepares a create request whose plan mirrors a real create:
// every unconfigured Optional+Computed attribute is unknown in the plan.
// When applyErrorCodeDefault is set, the schema's error_code_if_disabled
// default is applied the way the plan engine does at plan time.
func runAliasCreate(t *testing.T, apiURL string, applyErrorCodeDefault bool) *resource.CreateResponse {
	t.Helper()

	client, err := forwardemail.NewClient("test-key", forwardemail.WithAPIURL(apiURL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	r := &aliasResource{client: client}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", schemaResp.Diagnostics)
	}

	planModel := AliasModel{
		ID:                    types.StringUnknown(),
		Domain:                types.StringValue("akbnet.org"),
		Name:                  types.StringValue("postmaster"),
		Enabled:               types.BoolUnknown(),
		RecipientVerification: types.BoolUnknown(),
		HasIMAP:               types.BoolUnknown(),
		HasPGP:                types.BoolUnknown(),
		PublicKey:             types.StringUnknown(),
		HasSMIME:              types.BoolUnknown(),
		SMIMECertificate:      types.StringUnknown(),
		HasWKDDisabled:        types.BoolUnknown(),
		ErrorCodeIfDisabled:   types.Int64Unknown(),
		MaxQuota:              types.StringUnknown(),
		Retention:             types.Int64Unknown(),
		SMTPLimit:             types.Int64Unknown(),
		Recipients:            types.ListValueMust(types.StringType, []attr.Value{types.StringValue("user@external-example.com")}),
		Description:           types.StringUnknown(),
		Labels:                types.ListNull(types.StringType),
	}

	if applyErrorCodeDefault {
		attrSchema, ok := schemaResp.Schema.Attributes["error_code_if_disabled"]
		if !ok {
			t.Fatalf("attribute error_code_if_disabled missing from schema")
		}
		int64Attr, ok := attrSchema.(schema.Int64Attribute)
		if !ok {
			t.Fatalf("attribute error_code_if_disabled is not an Int64Attribute")
		}
		if int64Attr.Default == nil {
			t.Fatalf("error_code_if_disabled has no default; want StaticInt64(250)")
		}
		defResp := &defaults.Int64Response{PlanValue: types.Int64Unknown()}
		int64Attr.Default.DefaultInt64(
			context.Background(),
			defaults.Int64Request{Path: path.Root("error_code_if_disabled")},
			defResp,
		)
		if defResp.Diagnostics.HasError() {
			t.Fatalf("DefaultInt64 diagnostics: %v", defResp.Diagnostics)
		}
		planModel.ErrorCodeIfDisabled = defResp.PlanValue
	}

	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := plan.Set(context.Background(), planModel); diags.HasError() {
		t.Fatalf("plan.Set: %v", diags)
	}

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Create diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

// Attributes the user never configures are unknown in the create plan. The
// framework's Value*Pointer() helpers fabricate zero values for unknowns,
// which made Create send error_code_if_disabled=0 (rejected by the API:
// "Error code if disabled must be either 250, 421, or 550") as well as
// empty public_key/max_quota/retention/etc. Create must omit every
// unconfigured parameter instead of inventing values for them.
func TestAliasCreateOmitsUnconfiguredParams(t *testing.T) {
	svr, gotForm := createAliasTestServer(t)
	runAliasCreate(t, svr.URL, false)

	for _, key := range []string{
		"description",
		"public_key",
		"smime_certificate",
		"max_quota",
		"has_recipient_verification",
		"has_imap",
		"has_pgp",
		"has_smime",
		"has_wkd_disabled",
		"error_code_if_disabled",
		"retention",
		"smtp_limit",
		"labels[]",
	} {
		if gotForm.Has(key) {
			t.Errorf("create must not send unconfigured %q; got %v", key, *gotForm)
		}
	}
}

// error_code_if_disabled is one-way (not reliably echoed by the API), so
// when the user does not configure it the schema default (the Forward
// Email default error code 250) is what the plan carries into the request,
// and state must end up with 250.
func TestAliasCreateDefaultsErrorCodeIfDisabledTo250(t *testing.T) {
	svr, gotForm := createAliasTestServer(t)
	resp := runAliasCreate(t, svr.URL, true)

	if got := gotForm.Get("error_code_if_disabled"); got != "250" {
		t.Errorf("error_code_if_disabled = %q, want 250 (unconfigured must default to the Forward Email default, not 0)", got)
	}

	var final AliasModel
	if diags := resp.State.Get(context.Background(), &final); diags.HasError() {
		t.Fatalf("state.Get: %v", diags)
	}
	if got := final.ErrorCodeIfDisabled.ValueInt64(); got != 250 {
		t.Errorf("final error_code_if_disabled = %d, want 250", got)
	}
}
