package forwardemail

import (
	"context"
	"fmt"
	"strings"

	"github.com/forwardemail/forwardemail-api-go/forwardemail"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &aliasResource{}
	_ resource.ResourceWithConfigure = &aliasResource{}
)

type aliasResource struct {
	client *forwardemail.Client
}

func NewAliasResource() resource.Resource {
	return &aliasResource{}
}

type AliasModel struct {
	Id                    types.String `tfsdk:"id"`
	Domain                types.String `tfsdk:"domain"`
	Name                  types.String `tfsdk:"name"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	RecipientVerification types.Bool   `tfsdk:"recipient_verification"`
	Recipient             types.List   `tfsdk:"recipients"`
	Description           types.String `tfsdk:"description"`
	Labels                types.List   `tfsdk:"labels"`
}

func (aliasResource *aliasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alias"
}

func (aliasResource *aliasResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A resource to create Forward Email domain aliases.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Alias name (the ID is the same as the name).",
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "Fully qualified domain name (FQDN).",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Alias name.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether to enable to disable this alias.",
			},
			"recipient_verification": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to enable to require recipients to click an email verification link for emails to flow through.",
			},
			"recipients": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				Description: "List of recipients as valid email addresses, fully-qualified domain names (FQDN), IP addresses, or webhook URL's.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Alias description.",
			},
			"labels": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				Description: "List of labels.",
			},
		},
	}
}

func (aliasResource *aliasResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*forwardemail.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *forwardemail.Client, got: %T. Please report this issue to the provider maintainers.", req.ProviderData),
		)
		return
	}
	aliasResource.client = client
}

func (aliasResource *aliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AliasModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()
	aliasName := plan.Name.ValueString()

	var recipients, labels []string
	resp.Diagnostics.Append(plan.Recipient.ElementsAs(ctx, &recipients, false)...)
	resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := forwardemail.AliasParameters{
		HasRecipientVerification: plan.RecipientVerification.ValueBoolPointer(),
		IsEnabled:                plan.Enabled.ValueBoolPointer(),
		Recipients:               &recipients,
		Labels:                   &labels,
		Description:              plan.Description.ValueString(),
	}

	alias, err := aliasResource.client.CreateAlias(domain, aliasName, params)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create alias", err.Error())
		return
	}

	// N.B.: the SDKv2 provider stored the alias *name* as the ID, even
	// though the API returns its own alias ID. Keep that for state parity.
	plan.Id = types.StringValue(aliasName)
	plan.Domain = types.StringValue(alias.Domain.Name)
	plan.Enabled = types.BoolValue(alias.IsEnabled)
	plan.RecipientVerification = types.BoolValue(alias.HasRecipientVerification)
	plan.Description = types.StringValue(alias.Description)

	recipientsList, diags := types.ListValueFrom(ctx, types.StringType, alias.Recipients)
	resp.Diagnostics.Append(diags...)
	plan.Recipient = recipientsList

	labelsList, diags := types.ListValueFrom(ctx, types.StringType, alias.Labels)
	resp.Diagnostics.Append(diags...)
	plan.Labels = labelsList

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (aliasResource *aliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AliasModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	aliasName := state.Name.ValueString()

	alias, err := aliasResource.client.GetAlias(domain, aliasName)
	if err != nil {
		// The client returns a plain error embedding the status code
		// ("status: %d, body: %s"); there is no typed APIError in this
		// version, so match the string.
		if strings.Contains(err.Error(), "status: 404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read alias", err.Error())
		return
	}

	state.Id = types.StringValue(aliasName)
	state.Domain = types.StringValue(alias.Domain.Name)
	state.Enabled = types.BoolValue(alias.IsEnabled)
	state.RecipientVerification = types.BoolValue(alias.HasRecipientVerification)
	state.Description = types.StringValue(alias.Description)

	recipientsList, diags := types.ListValueFrom(ctx, types.StringType, alias.Recipients)
	resp.Diagnostics.Append(diags...)
	state.Recipient = recipientsList

	labelsList, diags := types.ListValueFrom(ctx, types.StringType, alias.Labels)
	resp.Diagnostics.Append(diags...)
	state.Labels = labelsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (aliasResource *aliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state AliasModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var recipients, labels []string
	resp.Diagnostics.Append(plan.Recipient.ElementsAs(ctx, &recipients, false)...)
	resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// N.B.: the Forward Email API does not behave correctly with partial
	// updates, so instead of sending only the changed parameters we must
	// send all of them.
	params := forwardemail.AliasParameters{
		HasRecipientVerification: plan.RecipientVerification.ValueBoolPointer(),
		IsEnabled:                plan.Enabled.ValueBoolPointer(),
		Recipients:               &recipients,
		Labels:                   &labels,
		Description:              plan.Description.ValueString(),
	}

	_, err := aliasResource.client.UpdateAlias(state.Domain.ValueString(), state.Id.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update alias", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (aliasResource *aliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AliasModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	aliasName := state.Id.ValueString()

	err := aliasResource.client.DeleteAlias(domain, aliasName)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete alias", err.Error())
		return
	}
}
