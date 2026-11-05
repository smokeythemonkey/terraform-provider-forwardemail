package forwardemail

import "github.com/hashicorp/terraform-plugin-framework/attr"

// changed reports whether a plan value differs from its prior state value,
// the framework equivalent of the SDKv2 d.HasChange helper. Equal treats
// two nulls and two unknowns as unchanged.
func changed(plan, state attr.Value) bool {
	return !plan.Equal(state)
}
