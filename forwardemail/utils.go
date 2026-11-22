package forwardemail

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func changed(plan, state attr.Value) bool {
	return !plan.Equal(state)
}

// The types.Value*Pointer() helpers return a pointer to the zero value for
// an unknown value, which fabricates parameters the user never configured
// (e.g. error_code_if_disabled=0, rejected by the API with "Error code if
// disabled must be either 250, 421, or 550"). Known*Pointer only returns
// a pointer for null or known values, so attributes that are still
// unknown in the plan are omitted from the request.
func knownBoolPointer(v types.Bool) *bool {
	if v.IsUnknown() {
		return nil
	}
	return v.ValueBoolPointer()
}

func knownInt64Pointer(v types.Int64) *int64 {
	if v.IsUnknown() {
		return nil
	}
	return v.ValueInt64Pointer()
}

func knownStringPointer(v types.String) *string {
	if v.IsUnknown() {
		return nil
	}
	return v.ValueStringPointer()
}
