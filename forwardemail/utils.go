package forwardemail

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// changed reports whether a plan value differs from its prior state value,
// the framework equivalent of the SDKv2 d.HasChange helper. Equal treats
// two nulls and two unknowns as unchanged.
func changed(plan, state attr.Value) bool {
	return !plan.Equal(state)
}

// stringValueOrNull returns a null String when s is empty. One-way API
// params that are Optional-only must stay null in state when unset, or
// Terraform reports "inconsistent result after apply" (null plan vs "" state).
func stringValueOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// int64ValueOrNull returns a null Int64 when n is zero. Same reasoning as
// stringValueOrNull for Optional-only numeric params.
func int64ValueOrNull(n int64) types.Int64 {
	if n == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(n)
}
