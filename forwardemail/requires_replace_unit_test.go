package forwardemail

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// requiresReplacePresent reports whether any plan modifier actually sets
// RequiresReplace for a changed value. RequiresReplace itself bails unless
// State and Plan raw values are non-null (resource update), so the request
// must carry those.
func requiresReplacePresent(mods []planmodifier.String) bool {
	nonNull := tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"x": tftypes.String}}, map[string]tftypes.Value{"x": tftypes.NewValue(tftypes.String, "v")})
	state := tfsdk.State{Raw: nonNull}
	plan := tfsdk.Plan{Raw: nonNull}
	for _, m := range mods {
		resp := &planmodifier.StringResponse{RequiresReplace: false}
		m.PlanModifyString(context.Background(), planmodifier.StringRequest{
			Plan:       plan,
			State:      state,
			PlanValue:  types.StringValue("new"),
			StateValue: types.StringValue("old"),
		}, resp)
		if resp.RequiresReplace {
			return true
		}
	}
	return false
}

// name (and domain on the alias) is the resource identity: every API call
// and the import ID are keyed by it, and the API cannot rename. Without
// RequiresReplace, a rename runs as an in-place update that sends the old
// name and fails with "inconsistent result after apply". These attributes
// must force destroy+recreate.
func TestIdentityAttributesRequireReplace(t *testing.T) {
	tests := []struct {
		name        string
		newResource func() resource.Resource
		attrs       []string
	}{
		{"domain", func() resource.Resource { return &domainResource{} }, []string{"name"}},
		{"alias", func() resource.Resource { return &aliasResource{} }, []string{"domain", "name"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &resource.SchemaResponse{}
			tt.newResource().Schema(context.Background(), resource.SchemaRequest{}, resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
			}
			for _, attrName := range tt.attrs {
				attrSchema, ok := resp.Schema.Attributes[attrName]
				if !ok {
					t.Fatalf("attribute %q missing", attrName)
				}
				strAttr, ok := attrSchema.(schema.StringAttribute)
				if !ok {
					t.Fatalf("attribute %q is not a StringAttribute", attrName)
				}
				if !requiresReplacePresent(strAttr.PlanModifiers) {
					t.Errorf("%s.%s must force replacement on change (RequiresReplace), got %d plan modifiers, none requiring replace", tt.name, attrName, len(strAttr.PlanModifiers))
				}
			}
		})
	}
}

var _ = stringplanmodifier.RequiresReplace
