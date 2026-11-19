package forwardemail

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Update must leave every non-identity attribute known after apply. When the
// user never configured an Optional+Computed attribute, the plan value is
// unknown; carrying the prior state value keeps the result fully known and
// avoids "provider returned invalid result object after apply".
func TestAliasUpdateFallbackFillsUnknownPlanValues(t *testing.T) {
	plan := AliasModel{
		ID:                    types.StringUnknown(),
		Domain:                types.StringValue("stark.com"),
		Name:                  types.StringValue("tony"),
		Enabled:               types.BoolValue(true),
		RecipientVerification: types.BoolUnknown(),
		HasIMAP:               types.BoolValue(true),
		HasPGP:                types.BoolUnknown(),
		PublicKey:             types.StringUnknown(),
		HasSMIME:              types.BoolUnknown(),
		SMIMECertificate:      types.StringUnknown(),
		HasWKDDisabled:        types.BoolUnknown(),
		ErrorCodeIfDisabled:   types.Int64Unknown(),
		MaxQuota:              types.StringUnknown(),
		Retention:             types.Int64Unknown(),
		SMTPLimit:             types.Int64Unknown(),
		Recipients:            types.ListUnknown(types.StringType),
		Description:           types.StringUnknown(),
		Labels:                types.ListUnknown(types.StringType),
	}
	state := AliasModel{
		Enabled:               types.BoolValue(false),
		RecipientVerification: types.BoolValue(true),
		HasIMAP:               types.BoolValue(false),
		HasPGP:                types.BoolValue(true),
		PublicKey:             types.StringValue("key"),
		HasSMIME:              types.BoolValue(true),
		SMIMECertificate:      types.StringValue("cert"),
		HasWKDDisabled:        types.BoolValue(false),
		ErrorCodeIfDisabled:   types.Int64Value(421),
		MaxQuota:              types.StringValue("1 GB"),
		Retention:             types.Int64Value(30),
		SMTPLimit:             types.Int64Value(100),
		Recipients:            types.ListValueMust(types.StringType, nil),
		Description:           types.StringValue("desc"),
		Labels:                types.ListValueMust(types.StringType, nil),
	}

	fillUnknownPlanFromState(&plan, &state)

	want := map[string]attr.Value{
		"RecipientVerification": types.BoolValue(true),
		"HasPGP":                types.BoolValue(true),
		"PublicKey":             types.StringValue("key"),
		"HasSMIME":              types.BoolValue(true),
		"SMIMECertificate":      types.StringValue("cert"),
		"HasWKDDisabled":        types.BoolValue(false),
		"ErrorCodeIfDisabled":   types.Int64Value(421),
		"MaxQuota":              types.StringValue("1 GB"),
		"Retention":             types.Int64Value(30),
		"SMTPLimit":             types.Int64Value(100),
		"Recipients":            state.Recipients,
		"Description":           types.StringValue("desc"),
		"Labels":                state.Labels,
	}
	got := map[string]attr.Value{
		"RecipientVerification": plan.RecipientVerification,
		"HasPGP":                plan.HasPGP,
		"PublicKey":             plan.PublicKey,
		"HasSMIME":              plan.HasSMIME,
		"SMIMECertificate":      plan.SMIMECertificate,
		"HasWKDDisabled":        plan.HasWKDDisabled,
		"ErrorCodeIfDisabled":   plan.ErrorCodeIfDisabled,
		"MaxQuota":              plan.MaxQuota,
		"Retention":             plan.Retention,
		"SMTPLimit":             plan.SMTPLimit,
		"Recipients":            plan.Recipients,
		"Description":           plan.Description,
		"Labels":                plan.Labels,
	}
	for name, wantValue := range want {
		if gotValue := got[name]; gotValue.IsUnknown() || !gotValue.Equal(wantValue) {
			t.Errorf("%s = %v, want known %v from state", name, gotValue, wantValue)
		}
	}

	// Known plan values must be preserved, not overwritten by state.
	if !plan.Enabled.ValueBool() {
		t.Errorf("Enabled = %v, want planned true preserved", plan.Enabled)
	}
	if !plan.HasIMAP.ValueBool() {
		t.Errorf("HasIMAP = %v, want planned true preserved", plan.HasIMAP)
	}
}
