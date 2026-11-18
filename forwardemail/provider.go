// Package forwardemail is a terraform provider for Forward Email
package forwardemail

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/smokeythemonkey/terraform-provider-forwardemail/internal/forwardemail"
)

// providerConfig is passed to resources via Configure as ProviderData.
type providerConfig struct {
	client *forwardemail.Client
	// allowDomainDeletion gates forwardemail_domain Delete. Deleting a
	// domain in Forward Email is destructive and irreversible, so the
	// provider refuses unless this is explicitly enabled.
	allowDomainDeletion bool
}

type forwardemailProvider struct {
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &forwardemailProvider{version: version}
	}
}

func (p *forwardemailProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "forwardemail"
	resp.Version = p.version
}

func (p *forwardemailProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Terraform provider for Forward Email.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key for Forwardemail access and management.",
				Optional:            true,
				Sensitive:           true,
			},
			"allow_domain_deletion": schema.BoolAttribute{
				MarkdownDescription: "Allow `forwardemail_domain` resources to be deleted. Domain deletion in Forward Email is destructive and irreversible, so it is refused unless this is set to `true` (or `FORWARDEMAIL_ALLOW_DOMAIN_DELETION=true`).",
				Optional:            true,
			},
		},
	}
}

func (p *forwardemailProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var apiKey types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("api_key"), &apiKey)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if apiKey.IsNull() {
		apiKey = types.StringValue(os.Getenv("FORWARDEMAIL_API_KEY"))
	}

	if apiKey.IsNull() || apiKey.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing Forward Email API Key",
			"Set the api_key provider attribute or the FORWARDEMAIL_API_KEY environment variable.",
		)
		return
	}

	var allowDomainDeletion types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("allow_domain_deletion"), &allowDomainDeletion)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if allowDomainDeletion.IsNull() {
		allowDomainDeletion = types.BoolValue(os.Getenv("FORWARDEMAIL_ALLOW_DOMAIN_DELETION") == "true")
	}

	client, err := forwardemail.NewClient(apiKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create Forward Email client", err.Error(),
		)
		return
	}
	cfg := &providerConfig{client: client, allowDomainDeletion: allowDomainDeletion.ValueBool()}
	resp.ResourceData = cfg
	resp.DataSourceData = cfg
}

func (p *forwardemailProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDomainResource,
		NewAliasResource,
	}
}

func (p *forwardemailProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAccountDataSource,
	}
}
