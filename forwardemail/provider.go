// Package forwardemail is a terraform provider for Forward Email
package forwardemail

import (
	"context"
	"os"

	"github.com/forwardemail/forwardemail-api-go/forwardemail"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
		},
	}
}

func (p *forwardemailProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var apiKey types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("api_key"), &apiKey)...)

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

	client := forwardemail.NewClient(forwardemail.ClientOptions{ApiKey: apiKey.ValueString()})
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *forwardemailProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDomainResource,
	}
}

func (p *forwardemailProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAccountDataSource,
	}
}
