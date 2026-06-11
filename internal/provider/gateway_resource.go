package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &gatewayResource{}
	_ resource.ResourceWithImportState = &gatewayResource{}
)

type gatewayResource struct{ client *Client }

func NewGatewayResource() resource.Resource { return &gatewayResource{} }

type gatewayModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Profile   types.String `tfsdk:"profile"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
	Realm     types.String `tfsdk:"realm"`
	Proxy     types.String `tfsdk:"proxy"`
	Register  types.Bool   `tfsdk:"register"`
	Params    types.Map    `tfsdk:"params"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *gatewayResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway"
}

func (r *gatewayResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A FreeSWITCH Sofia SIP gateway / trunk.",
		Attributes: map[string]schema.Attribute{
			"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":     schema.StringAttribute{Required: true, PlanModifiers: replace},
			"profile":  schema.StringAttribute{Required: true, PlanModifiers: replace},
			"enabled":  schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"username": schema.StringAttribute{Optional: true},
			"password": schema.StringAttribute{Optional: true, Sensitive: true},
			"realm":    schema.StringAttribute{Optional: true},
			"proxy":    schema.StringAttribute{Required: true},
			"register": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"params":   schema.MapAttribute{Optional: true, ElementType: types.StringType},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *gatewayResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *gatewayResource) apiFromModel(ctx context.Context, m gatewayModel) apiGateway {
	params, _ := tfMapToGo(ctx, m.Params)
	return apiGateway{
		Name:     m.Name.ValueString(),
		Profile:  m.Profile.ValueString(),
		Enabled:  m.Enabled.ValueBool(),
		Username: m.Username.ValueString(),
		Password: m.Password.ValueString(),
		Realm:    m.Realm.ValueString(),
		Proxy:    m.Proxy.ValueString(),
		Register: m.Register.ValueBool(),
		Params:   params,
	}
}

func optStr(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func (r *gatewayResource) modelFromAPI(ctx context.Context, g *apiGateway, m *gatewayModel) error {
	m.ID = types.StringValue(g.Profile + "/" + g.Name)
	m.Name = types.StringValue(g.Name)
	m.Profile = types.StringValue(g.Profile)
	m.Enabled = types.BoolValue(g.Enabled)
	m.Username = optStr(g.Username)
	m.Password = optStr(g.Password)
	m.Realm = optStr(g.Realm)
	m.Proxy = types.StringValue(g.Proxy)
	m.Register = types.BoolValue(g.Register)
	params, diags := goMapToTFPreserveNull(ctx, g.Params, m.Params)
	if diags.HasError() {
		return errors.New("params conversion failed")
	}
	m.Params = params
	m.CreatedAt = types.StringValue(g.CreatedAt)
	m.UpdatedAt = types.StringValue(g.UpdatedAt)
	return nil
}

func (r *gatewayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m gatewayModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.createGateway(ctx, r.apiFromModel(ctx, m))
	if err != nil {
		resp.Diagnostics.AddError("create gateway failed", err.Error())
		return
	}
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *gatewayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m gatewayModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.getGateway(ctx, m.Profile.ValueString(), m.Name.ValueString())
	if errors.Is(err, ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("read gateway failed", err.Error())
		return
	}
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *gatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m gatewayModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.updateGateway(ctx, m.Profile.ValueString(), m.Name.ValueString(), r.apiFromModel(ctx, m))
	if err != nil {
		resp.Diagnostics.AddError("update gateway failed", err.Error())
		return
	}
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *gatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m gatewayModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.deleteGateway(ctx, m.Profile.ValueString(), m.Name.ValueString())
	if err != nil && !errors.Is(err, ErrNotFound) {
		resp.Diagnostics.AddError("delete gateway failed", err.Error())
	}
}

func (r *gatewayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("invalid import id", fmt.Sprintf("expected \"profile/name\", got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("profile"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}
