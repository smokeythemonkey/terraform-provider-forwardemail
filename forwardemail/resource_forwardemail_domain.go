package forwardemail

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/forwardemail/forwardemail-api-go/forwardemail"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &domainResource{}
	_ resource.ResourceWithConfigure = &domainResource{}
)

type domainResource struct {
	client *forwardemail.Client
}

func NewDomainResource() resource.Resource {
	return &domainResource{}
}

type DomainModel struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	AdultContentProtection types.Bool   `tfsdk:"adult_content_protection"`
	PhishingProtection     types.Bool   `tfsdk:"phishing_protection"`
	ExecutableProtection   types.Bool   `tfsdk:"executable_protection"`
	VirusProtection        types.Bool   `tfsdk:"virus_protection"`
	RecipientVerification  types.Bool   `tfsdk:"recipient_verification"`
}

func (domainResource *domainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (domainResource *domainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A resource to create Forward Email domains.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Domain name (the ID is the same as the name).",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Fully qualified domain name (FQDN) or IP address.",
			},
			"adult_content_protection": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether to enable Spam Scanner adult content protection on this domain.",
			},
			"phishing_protection": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether to enable Spam Scanner phishing protection on this domain.",
			},
			"executable_protection": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether to enable Spam Scanner executable protection on this domain.",
			},
			"virus_protection": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether to enable Spam Scanner virus protection on this domain.",
			},
			"recipient_verification": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Global domain default for whether to require alias recipients to click an email verification link for emails to flow through.",
			},
		},
	}
}

func (domainResource *domainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	domainResource.client = client
}

func (domainResource *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DomainModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := forwardemail.DomainParameters{
		HasAdultContentProtection: plan.AdultContentProtection.ValueBoolPointer(),
		HasPhishingProtection:     plan.PhishingProtection.ValueBoolPointer(),
		HasExecutableProtection:   plan.ExecutableProtection.ValueBoolPointer(),
		HasVirusProtection:        plan.VirusProtection.ValueBoolPointer(),
		HasRecipientVerification:  plan.RecipientVerification.ValueBoolPointer(),
	}

	domain, err := domainResource.client.CreateDomain(ctx, plan.Name.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create domain", err.Error())
		return
	}

	plan.Id = types.StringValue(domain.ID)
	plan.Name = types.StringValue(domain.Name)
	plan.AdultContentProtection = types.BoolValue(domain.HasAdultContentProtection)
	plan.PhishingProtection = types.BoolValue(domain.HasPhishingProtection)
	plan.ExecutableProtection = types.BoolValue(domain.HasExecutableProtection)
	plan.VirusProtection = types.BoolValue(domain.HasVirusProtection)
	plan.RecipientVerification = types.BoolValue(domain.HasRecipientVerification)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (domainResource *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainModel
	var apiErr *forwardemail.APIError

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := domainResource.client.GetDomain(ctx, state.Id.ValueString())
	if err != nil {
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read domain", err.Error())
		return
	}

	state.Id = types.StringValue(domain.ID)
	state.Name = types.StringValue(domain.Name)
	state.AdultContentProtection = types.BoolValue(domain.HasAdultContentProtection)
	state.PhishingProtection = types.BoolValue(domain.HasPhishingProtection)
	state.ExecutableProtection = types.BoolValue(domain.HasExecutableProtection)
	state.VirusProtection = types.BoolValue(domain.HasVirusProtection)
	state.RecipientVerification = types.BoolValue(domain.HasRecipientVerification)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (domainResource *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state DomainModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := forwardemail.DomainParameters{}

	if changed(plan.AdultContentProtection, state.AdultContentProtection) {
		params.HasAdultContentProtection = plan.AdultContentProtection.ValueBoolPointer()
	}
	if changed(plan.PhishingProtection, state.PhishingProtection) {
		params.HasPhishingProtection = plan.PhishingProtection.ValueBoolPointer()
	}
	if changed(plan.ExecutableProtection, state.ExecutableProtection) {
		params.HasExecutableProtection = plan.ExecutableProtection.ValueBoolPointer()
	}
	if changed(plan.VirusProtection, state.VirusProtection) {
		params.HasVirusProtection = plan.VirusProtection.ValueBoolPointer()
	}
	if changed(plan.RecipientVerification, state.RecipientVerification) {
		params.HasRecipientVerification = plan.RecipientVerification.ValueBoolPointer()
	}

	_, err := domainResource.client.UpdateDomain(ctx, state.Id.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update domain", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (domainResource *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Id.ValueString()
	err := domainResource.client.DeleteDomain(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete Domain", err.Error())
		return
	}
}
