package provider

import (
	"context"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type freeswitchProvider struct {
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &freeswitchProvider{version: version}
	}
}

type providerModel struct {
	Endpoint   types.String `tfsdk:"endpoint"`
	Token      types.String `tfsdk:"token"`
	CACertFile types.String `tfsdk:"ca_cert_file"`
	Insecure   types.Bool   `tfsdk:"insecure"`
}

func (p *freeswitchProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "freeswitch"
	resp.Version = p.version
}

func (p *freeswitchProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage FreeSWITCH through the FreeSWITCH IaC control-plane API. " +
			"This provider is a client for that API and requires a running control-plane " +
			"(see https://github.com/jamalshahverdiev/freeswitch-iac-platform); it does not talk to FreeSWITCH directly.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Control-plane base URL. Env: `FREESWITCH_ENDPOINT`. Example `https://localhost:8080`.",
			},
			"token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Bearer token for `/api/v1`. Env: `FREESWITCH_TOKEN`.",
			},
			"ca_cert_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a CA cert (PEM) to verify the control-plane TLS cert. Env: `FREESWITCH_CACERT`.",
			},
			"insecure": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Skip TLS verification. Env: `FREESWITCH_INSECURE`. Not for production.",
			},
		},
	}
}

func (p *freeswitchProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := firstNonEmpty(cfg.Endpoint.ValueString(), os.Getenv("FREESWITCH_ENDPOINT"), "https://localhost:8080")
	token := firstNonEmpty(cfg.Token.ValueString(), os.Getenv("FREESWITCH_TOKEN"), "")
	caCert := firstNonEmpty(cfg.CACertFile.ValueString(), os.Getenv("FREESWITCH_CACERT"), "")
	insecure := cfg.Insecure.ValueBool()
	if !insecure {
		if v, err := strconv.ParseBool(os.Getenv("FREESWITCH_INSECURE")); err == nil {
			insecure = v
		}
	}

	if token == "" {
		resp.Diagnostics.AddError("missing token", "Set provider `token` or FREESWITCH_TOKEN.")
		return
	}

	client, err := NewClient(endpoint, token, caCert, insecure)
	if err != nil {
		resp.Diagnostics.AddError("client init failed", err.Error())
		return
	}
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *freeswitchProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDomainResource,
		NewUserResource,
		NewGatewayResource,
		NewDialplanExtensionResource,
		NewReloadXMLResource,
		NewCCQueueResource,
		NewCCAgentResource,
		NewCCTierResource,
		NewCCReloadResource,
		NewConfProfileResource,
		NewConfRoomResource,
	}
}

func (p *freeswitchProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDomainDataSource,
		NewUserDataSource,
		NewGatewayDataSource,
		NewDialplanExtensionDataSource,
		NewGatewayStatusDataSource,
		NewUserRegistrationDataSource,
		NewConferenceStatusDataSource,
		NewCCQueueDataSource,
		NewCCAgentDataSource,
		NewCCTierDataSource,
		NewConfProfileDataSource,
		NewConfRoomDataSource,
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// clientFromProviderData is a helper for resources' Configure.
func clientFromProviderData(data any, diags interface{ AddError(string, string) }) *Client {
	if data == nil {
		return nil
	}
	c, ok := data.(*Client)
	if !ok {
		diags.AddError("unexpected provider data", "expected *Client")
		return nil
	}
	return c
}
