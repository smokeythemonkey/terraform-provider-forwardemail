package forwardemail

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/smokeythemonkey/terraform-provider-forwardemail/internal/forwardemail"
)

var (
	_ resource.Resource                = &aliasResource{}
	_ resource.ResourceWithConfigure   = &aliasResource{}
	_ resource.ResourceWithImportState = &aliasResource{}
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
	HasIMAP               types.Bool   `tfsdk:"has_imap"`
	HasPGP                types.Bool   `tfsdk:"has_pgp"`
	PublicKey             types.String `tfsdk:"public_key"`
	HasSMIME              types.Bool   `tfsdk:"has_smime"`
	SMIMECertificate      types.String `tfsdk:"smime_certificate"`
	HasWKDDisabled        types.Bool   `tfsdk:"has_wkd_disabled"`
	ErrorCodeIfDisabled   types.Int64  `tfsdk:"error_code_if_disabled"`
	MaxQuota              types.String `tfsdk:"max_quota"`
	Retention             types.Int64  `tfsdk:"retention"`
	SMTPLimit             types.Int64  `tfsdk:"smtp_limit"`
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
			"has_imap": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to enable the webmail (IMAP) option for this alias.",
			},
			"has_pgp": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to enable PGP encryption for this alias.",
			},
			"public_key": schema.StringAttribute{
				Optional:    true,
				Description: "PGP public key (armored) used to encrypt inbound mail to this alias.",
			},
			"has_smime": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to enable S/MIME encryption for this alias.",
			},
			"smime_certificate": schema.StringAttribute{
				Optional:    true,
				Description: "S/MIME certificate (PEM) used to encrypt inbound mail to this alias.",
			},
			"has_wkd_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to disable Web Key Directory (WKD) lookups for this alias.",
			},
			"error_code_if_disabled": schema.Int64Attribute{
				Optional:    true,
				Description: "SMTP response code to use when the alias is disabled (one of 250, 421, 550).",
			},
			"max_quota": schema.StringAttribute{
				Optional:    true,
				Description: "Maximum storage quota for this alias (e.g. \"1 GB\").",
			},
			"retention": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of days to retain emails for this alias (0-365).",
			},
			"smtp_limit": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximum number of outbound SMTP messages per day for this alias (0 for unlimited).",
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

	cfg, ok := req.ProviderData.(*providerConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *providerConfig, got: %T. Please report this issue to the provider maintainers.", req.ProviderData),
		)
		return
	}
	aliasResource.client = cfg.client
}

func (aliasResource *aliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AliasModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var recipients, labels []string
	resp.Diagnostics.Append(plan.Recipient.ElementsAs(ctx, &recipients, false)...)
	resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := forwardemail.AliasParameters{
		HasRecipientVerification: plan.RecipientVerification.ValueBoolPointer(),
		IsEnabled:                plan.Enabled.ValueBoolPointer(),
		HasIMAP:                  plan.HasIMAP.ValueBoolPointer(),
		HasPGP:                   plan.HasPGP.ValueBoolPointer(),
		PublicKey:                plan.PublicKey.ValueStringPointer(),
		HasSMIME:                 plan.HasSMIME.ValueBoolPointer(),
		SMIMECertificate:         plan.SMIMECertificate.ValueStringPointer(),
		HasWKDDisabled:           plan.HasWKDDisabled.ValueBoolPointer(),
		ErrorCodeIfDisabled:      plan.ErrorCodeIfDisabled.ValueInt64Pointer(),
		MaxQuota:                 plan.MaxQuota.ValueStringPointer(),
		Retention:                plan.Retention.ValueInt64Pointer(),
		SMTPLimit:                plan.SMTPLimit.ValueInt64Pointer(),
		Recipients:               &recipients,
		Labels:                   &labels,
		Description:              plan.Description.ValueStringPointer(),
	}

	alias, err := aliasResource.client.CreateAlias(ctx, plan.Domain.ValueString(), plan.Name.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create alias", err.Error())
		return
	}

	// The API returns an opaque ID in alias.ID, but every client call is
	// keyed by the (domain, name) pair, so the name is the resource's
	// Terraform ID.
	plan.Id = types.StringValue(alias.Name)
	plan.Domain = types.StringValue(alias.Domain.Name)
	plan.Name = types.StringValue(alias.Name)
	// Echoed attributes are materialized from the API response (they are
	// Computed, so the plan value may be unknown here). One-way attributes
	// (public_key, smime_certificate, error_code_if_disabled, max_quota)
	// are NOT set from the response: the API may echo a default for them,
	// and overwriting a null plan value with it fails the apply with
	// "inconsistent result after apply".
	plan.Enabled = types.BoolValue(alias.IsEnabled)
	plan.RecipientVerification = types.BoolValue(alias.HasRecipientVerification)
	plan.HasIMAP = types.BoolValue(alias.HasIMAP)
	plan.HasPGP = types.BoolValue(alias.HasPGP)
	plan.HasSMIME = types.BoolValue(alias.HasSMIME)
	plan.HasWKDDisabled = types.BoolValue(alias.HasWKDDisabled)
	plan.Retention = types.Int64Value(int64(alias.Retention))
	plan.SMTPLimit = types.Int64Value(int64(alias.SMTPLimit))
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

	alias, err := aliasResource.client.GetAlias(ctx, state.Domain.ValueString(), state.Name.ValueString())
	if err != nil {
		if forwardemail.IsStatus(err, http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read alias", err.Error())
		return
	}

	state.Id = types.StringValue(alias.Name)
	state.Domain = types.StringValue(alias.Domain.Name)
	state.Name = types.StringValue(alias.Name)
	// Echoed attributes refresh from the API so out-of-band drift is
	// detected. One-way attributes (public_key, smime_certificate,
	// error_code_if_disabled, max_quota) are not reliably echoed and must
	// keep their prior state value.
	state.Enabled = types.BoolValue(alias.IsEnabled)
	state.RecipientVerification = types.BoolValue(alias.HasRecipientVerification)
	state.HasIMAP = types.BoolValue(alias.HasIMAP)
	state.HasPGP = types.BoolValue(alias.HasPGP)
	state.HasSMIME = types.BoolValue(alias.HasSMIME)
	state.HasWKDDisabled = types.BoolValue(alias.HasWKDDisabled)
	state.Retention = types.Int64Value(int64(alias.Retention))
	state.SMTPLimit = types.Int64Value(int64(alias.SMTPLimit))
	state.Description = types.StringValue(alias.Description)

	recipientsList, diags := types.ListValueFrom(ctx, types.StringType, alias.Recipients)
	resp.Diagnostics.Append(diags...)
	state.Recipient = recipientsList

	labelsList, diags := types.ListValueFrom(ctx, types.StringType, alias.Labels)
	resp.Diagnostics.Append(diags...)
	state.Labels = labelsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// changedAliasParams builds the API parameters for an update from only the
// attributes that differ between plan and state. The Forward Email alias
// API resets any absent field (description, labels, recipients) on PUT, so
// only the attributes that actually changed may be sent; sending everything
// would let the API clobber fields the user never configured.
func changedAliasParams(plan, state AliasModel, recipients, labels []string) forwardemail.AliasParameters {
	params := forwardemail.AliasParameters{}
	if changed(plan.RecipientVerification, state.RecipientVerification) {
		params.HasRecipientVerification = plan.RecipientVerification.ValueBoolPointer()
	}
	if changed(plan.Enabled, state.Enabled) {
		params.IsEnabled = plan.Enabled.ValueBoolPointer()
	}
	if changed(plan.HasIMAP, state.HasIMAP) {
		params.HasIMAP = plan.HasIMAP.ValueBoolPointer()
	}
	if changed(plan.HasPGP, state.HasPGP) {
		params.HasPGP = plan.HasPGP.ValueBoolPointer()
	}
	if changed(plan.PublicKey, state.PublicKey) {
		params.PublicKey = plan.PublicKey.ValueStringPointer()
	}
	if changed(plan.HasSMIME, state.HasSMIME) {
		params.HasSMIME = plan.HasSMIME.ValueBoolPointer()
	}
	if changed(plan.SMIMECertificate, state.SMIMECertificate) {
		params.SMIMECertificate = plan.SMIMECertificate.ValueStringPointer()
	}
	if changed(plan.HasWKDDisabled, state.HasWKDDisabled) {
		params.HasWKDDisabled = plan.HasWKDDisabled.ValueBoolPointer()
	}
	if changed(plan.ErrorCodeIfDisabled, state.ErrorCodeIfDisabled) {
		params.ErrorCodeIfDisabled = plan.ErrorCodeIfDisabled.ValueInt64Pointer()
	}
	if changed(plan.MaxQuota, state.MaxQuota) {
		params.MaxQuota = plan.MaxQuota.ValueStringPointer()
	}
	if changed(plan.Retention, state.Retention) {
		params.Retention = plan.Retention.ValueInt64Pointer()
	}
	if changed(plan.SMTPLimit, state.SMTPLimit) {
		params.SMTPLimit = plan.SMTPLimit.ValueInt64Pointer()
	}
	if changed(plan.Recipient, state.Recipient) {
		params.Recipients = &recipients
	}
	if changed(plan.Labels, state.Labels) {
		params.Labels = &labels
	}
	if changed(plan.Description, state.Description) {
		params.Description = plan.Description.ValueStringPointer()
	}
	return params
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

	params := changedAliasParams(plan, state, recipients, labels)

	alias, err := aliasResource.client.UpdateAlias(ctx, state.Domain.ValueString(), state.Name.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update alias", err.Error())
		return
	}

	// Only identity fields come from the API; every planned value is kept
	// so Terraform does not report an inconsistent result when the API
	// normalizes or does not echo a parameter back.
	plan.Id = types.StringValue(alias.Name)
	plan.Domain = types.StringValue(alias.Domain.Name)
	plan.Name = types.StringValue(alias.Name)

	if resp.Diagnostics.HasError() {
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

	err := aliasResource.client.DeleteAlias(ctx, state.Domain.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete alias", err.Error())
		return
	}
}

// ImportState accepts an import ID of "domain/name" because an alias is
// identified by that pair in the API. Read then fills every other attribute.
func (aliasResource *aliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.SplitN(req.ID, "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected 'domain/name', got: %q.", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
