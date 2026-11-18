package forwardemail

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Configure must stop on config-read diagnostics instead of proceeding to
// build the client with a zero value. A raw config missing the
// allow_domain_deletion attribute makes its GetAttribute read fail.
func TestProviderConfigureStopsOnConfigReadError(t *testing.T) {
	p := New("test")().(*forwardemailProvider)

	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", schemaResp.Diagnostics)
	}

	// Wrong type for allow_domain_deletion: the raw value declares it as a
	// number while the schema expects bool, so GetAttribute returns a
	// diagnostic. The object type must still declare both attributes so
	// construction doesn't panic; the mismatch is in the value itself.
	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"api_key":               tftypes.String,
				"allow_domain_deletion": tftypes.Bool,
			},
		}, map[string]tftypes.Value{
			"api_key":               tftypes.NewValue(tftypes.String, "test-key"),
			"allow_domain_deletion": tftypes.NewValue(tftypes.Bool, nil),
		}),
	}
	// Break the read by swapping in a raw value whose type disagrees.
	config.Raw = tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"api_key":               tftypes.String,
			"allow_domain_deletion": tftypes.String, // schema says bool
		},
	}, map[string]tftypes.Value{
		"api_key":               tftypes.NewValue(tftypes.String, "test-key"),
		"allow_domain_deletion": tftypes.NewValue(tftypes.String, "nope"),
	})

	resp := &provider.ConfigureResponse{}
	p.Configure(context.Background(), provider.ConfigureRequest{Config: config}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics from the failed allow_domain_deletion read, got none")
	}
	for _, d := range resp.Diagnostics.Errors() {
		if d.Summary() == "Failed to create Forward Email client" {
			t.Error("Configure proceeded past a failed config read and tried to build the client")
		}
	}
	if resp.ResourceData != nil {
		t.Error("ResourceData must not be set when a config read failed")
	}
}

// The deletion gate env var must accept the standard truthy spellings
// (true/TRUE/1), not just lowercase "true", and fail closed (false) on
// unparsable values.
func TestEnvAllowDomainDeletionParsing(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"1", true},
		{"false", false},
		{"0", false},
		{"garbage", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Setenv("FORWARDEMAIL_ALLOW_DOMAIN_DELETION", tt.value)
			t.Setenv("FORWARDEMAIL_API_KEY", "test-key")

			p := New("test")().(*forwardemailProvider)
			schemaResp := &provider.SchemaResponse{}
			p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

			config := tfsdk.Config{
				Schema: schemaResp.Schema,
				Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(context.Background()), map[string]tftypes.Value{
					"api_key":               tftypes.NewValue(tftypes.String, nil),
					"allow_domain_deletion": tftypes.NewValue(tftypes.Bool, nil),
				}),
			}

			resp := &provider.ConfigureResponse{}
			p.Configure(context.Background(), provider.ConfigureRequest{Config: config}, resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("Configure diagnostics: %v", resp.Diagnostics)
			}
			cfg, ok := resp.ResourceData.(*providerConfig)
			if !ok {
				t.Fatal("ResourceData is not *providerConfig")
			}
			if cfg.allowDomainDeletion != tt.want {
				t.Errorf("FORWARDEMAIL_ALLOW_DOMAIN_DELETION=%q -> allowDomainDeletion=%v, want %v", tt.value, cfg.allowDomainDeletion, tt.want)
			}
		})
	}
}
