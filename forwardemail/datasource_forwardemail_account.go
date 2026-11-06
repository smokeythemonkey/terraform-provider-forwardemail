package forwardemail

import (
	"context"

	"github.com/forwardemail/forwardemail-api-go/forwardemail"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &accountDataSource{}
	_ datasource.DataSourceWithConfigure = &accountDataSource{}
)

type accountDataSource struct {
	client *forwardemail.Client
}

func NewAccountDataSource() datasource.DataSource {
	return &accountDataSource{}
}

type AccountModel struct {
	Id          types.String `tfsdk:"id"`
	Plan        types.String `tfsdk:"plan"`
	Email       types.String `tfsdk:"email"`
	FullEmail   types.String `tfsdk:"full_email"`
	DisplayName types.String `tfsdk:"display_name"`
}

func (accountSource *accountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (accountSource *accountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A forwardemail datasource that represents the account properties",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Account ID.",
			},
			"plan": schema.StringAttribute{
				Computed:    true,
				Description: "Plan type.",
			},
			"email": schema.StringAttribute{
				Computed:    true,
				Description: "Email address.",
			},
			"full_email": schema.StringAttribute{
				Computed:    true,
				Description: "Full email address.",
			},
			"display_name": schema.StringAttribute{
				Computed:    true,
				Description: "Display name.",
			},
		},
	}
}

func (accountSource *accountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*forwardemail.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *forwardemail.Client, got something else. "+
				"Please report this issue to the provider maintainers",
		)
		return
	}
	accountSource.client = client
}

func (accountSource *accountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AccountModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	account, err := accountSource.client.GetAccount(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get account", err.Error())
		return
	}

	data.Id = types.StringValue(account.ID)
	data.Plan = types.StringValue(account.Plan)
	data.Email = types.StringValue(account.Email)
	data.FullEmail = types.StringValue(account.FullEmail)
	data.DisplayName = types.StringValue(account.DisplayName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
