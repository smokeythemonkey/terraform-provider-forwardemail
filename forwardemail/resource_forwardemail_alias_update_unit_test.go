package forwardemail

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/smokeythemonkey/terraform-provider-forwardemail/internal/forwardemail"
)

// aliasUpdateTestModel builds a fully populated AliasModel for update-path
// unit tests. The imap flag controls the has_imap value so a plan/state
// pair with a real change can be built (mirrors flipping has_imap in the
// config).
func aliasUpdateTestModel(imap bool) AliasModel {
	return AliasModel{
		ID:                    types.StringValue("purchases"),
		Domain:                types.StringValue("akbnet.org"),
		Name:                  types.StringValue("purchases"),
		Enabled:               types.BoolValue(true),
		RecipientVerification: types.BoolValue(false),
		HasIMAP:               types.BoolValue(imap),
		HasPGP:                types.BoolValue(true),
		PublicKey:             types.StringValue("pub"),
		HasSMIME:              types.BoolValue(false),
		SMIMECertificate:      types.StringNull(),
		HasWKDDisabled:        types.BoolValue(false),
		ErrorCodeIfDisabled:   types.Int64Value(421),
		MaxQuota:              types.StringValue("1 GB"),
		Retention:             types.Int64Value(30),
		SMTPLimit:             types.Int64Value(100),
		Recipients:            types.ListValueMust(types.StringType, []attr.Value{types.StringValue("user@external-example.com")}),
		Description:           types.StringValue("alias for purchases"),
		Labels:                types.ListValueMust(types.StringType, []attr.Value{types.StringValue("production")}),
	}
}

// The update-alias endpoint echoes "domain" as the domain's object ID (a hex
// string), not the FQDN. Update must keep the planned identity values instead
// of writing the response back into state, or the apply is rejected with
// "Provider produced inconsistent result after apply".
func TestAliasUpdateKeepsPlannedIdentityWhenResponseEchoesDomainID(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"user":"ops@akbnet.org","domain":"69b8e43aa18b37daca40258c","name":"purchases","is_enabled":true,"has_imap":false}`)
	}))
	defer svr.Close()

	client, err := forwardemail.NewClient("test-key", forwardemail.WithAPIURL(svr.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	r := &aliasResource{client: client}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", schemaResp.Diagnostics)
	}

	// During a real update, the Computed id is unknown in the plan; the
	// provider must fill it from the name.
	planModel := aliasUpdateTestModel(false)
	planModel.ID = types.StringUnknown()
	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := plan.Set(context.Background(), planModel); diags.HasError() {
		t.Fatalf("plan.Set: %v", diags)
	}
	priorState := aliasUpdateTestModel(true)
	state := tfsdk.State{Schema: schemaResp.Schema}
	if diags := state.Set(context.Background(), &priorState); diags.HasError() {
		t.Fatalf("state.Set: %v", diags)
	}

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Update(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Update diagnostics: %v", resp.Diagnostics)
	}

	var final AliasModel
	if diags := resp.State.Get(context.Background(), &final); diags.HasError() {
		t.Fatalf("state.Get: %v", diags)
	}

	if got := final.Domain.ValueString(); got != "akbnet.org" {
		t.Errorf("domain = %q, want %q (planned identity must survive the API echoing the domain object ID)", got, "akbnet.org")
	}
	if got := final.Name.ValueString(); got != "purchases" {
		t.Errorf("name = %q, want %q", got, "purchases")
	}
	if got := final.ID.ValueString(); got != "purchases" {
		t.Errorf("id = %q, want %q", got, "purchases")
	}
}
