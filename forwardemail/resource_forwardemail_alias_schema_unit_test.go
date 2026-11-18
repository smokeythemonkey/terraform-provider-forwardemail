package forwardemail

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The schema promises error_code_if_disabled is one of 250/421/550 and
// retention is 0-365; invalid values must fail at plan time with an
// attribute-scoped diagnostic, not only when the API rejects them.
func TestAliasSchemaValidators(t *testing.T) {
	r := &aliasResource{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", schemaResp.Diagnostics)
	}

	tests := []struct {
		name      string
		attr      string
		value     int64
		wantValid bool
	}{
		{"error code 250 valid", "error_code_if_disabled", 250, true},
		{"error code 421 valid", "error_code_if_disabled", 421, true},
		{"error code 550 valid", "error_code_if_disabled", 550, true},
		{"error code 999 invalid", "error_code_if_disabled", 999, false},
		{"retention 0 valid", "retention", 0, true},
		{"retention 365 valid", "retention", 365, true},
		{"retention 366 invalid", "retention", 366, false},
		{"retention negative invalid", "retention", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrSchema, ok := schemaResp.Schema.Attributes[tt.attr]
			if !ok {
				t.Fatalf("attribute %q missing from schema", tt.attr)
			}
			int64Attr, ok := attrSchema.(schema.Int64Attribute)
			if !ok {
				t.Fatalf("attribute %q is not an Int64Attribute", tt.attr)
			}
			req := validator.Int64Request{
				Path:        path.Root(tt.attr),
				ConfigValue: types.Int64Value(tt.value),
			}
			resp := &validator.Int64Response{}
			for _, v := range int64Attr.Validators {
				v.ValidateInt64(context.Background(), req, resp)
			}
			if gotValid := !resp.Diagnostics.HasError(); gotValid != tt.wantValid {
				t.Errorf("validators for %s=%d: got valid=%v, want %v (diags: %v)", tt.attr, tt.value, gotValid, tt.wantValid, resp.Diagnostics)
			}
		})
	}
}
