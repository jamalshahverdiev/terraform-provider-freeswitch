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
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

type userResource struct{ client *Client }

func NewUserResource() resource.Resource { return &userResource{} }

type userModel struct {
	ID        types.String `tfsdk:"id"`
	Domain    types.String `tfsdk:"domain"`
	Number    types.String `tfsdk:"number"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Params    types.Map    `tfsdk:"params"`
	Variables types.Map    `tfsdk:"variables"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A FreeSWITCH SIP user / extension.",
		Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"domain":  schema.StringAttribute{Required: true, PlanModifiers: replace},
			"number":  schema.StringAttribute{Required: true, PlanModifiers: replace},
			"enabled": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"params": schema.MapAttribute{
				Optional: true, Sensitive: true, ElementType: types.StringType,
				MarkdownDescription: "User params (e.g. password, vm-password). Sensitive.",
			},
			"variables":  schema.MapAttribute{Optional: true, ElementType: types.StringType},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *userResource) apiFromModel(ctx context.Context, m userModel) apiUser {
	params, _ := tfMapToGo(ctx, m.Params)
	vars, _ := tfMapToGo(ctx, m.Variables)
	return apiUser{
		Domain:    m.Domain.ValueString(),
		Number:    m.Number.ValueString(),
		Enabled:   m.Enabled.ValueBool(),
		Params:    params,
		Variables: vars,
	}
}

func (r *userResource) modelFromAPI(ctx context.Context, u *apiUser, m *userModel) error {
	m.ID = types.StringValue(u.Domain + "/" + u.Number)
	m.Domain = types.StringValue(u.Domain)
	m.Number = types.StringValue(u.Number)
	m.Enabled = types.BoolValue(u.Enabled)
	params, d1 := goMapToTFPreserveNull(ctx, u.Params, m.Params)
	vars, d2 := goMapToTFPreserveNull(ctx, u.Variables, m.Variables)
	if d1.HasError() || d2.HasError() {
		return errors.New("map conversion failed")
	}
	m.Params = params
	m.Variables = vars
	m.CreatedAt = types.StringValue(u.CreatedAt)
	m.UpdatedAt = types.StringValue(u.UpdatedAt)
	return nil
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.createUser(ctx, r.apiFromModel(ctx, m))
	if err != nil {
		resp.Diagnostics.AddError("create user failed", err.Error())
		return
	}
	// The API does not echo params back; keep what the plan set.
	planParams := m.Params
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	m.Params = planParams
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.getUser(ctx, m.Domain.ValueString(), m.Number.ValueString())
	if errors.Is(err, ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("read user failed", err.Error())
		return
	}
	stateParams := m.Params
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	// Preserve params from state if the API returned them (it may, but they are
	// not authoritative here); prefer existing state to avoid spurious drift.
	if !stateParams.IsNull() {
		m.Params = stateParams
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.updateUser(ctx, m.Domain.ValueString(), m.Number.ValueString(), r.apiFromModel(ctx, m))
	if err != nil {
		resp.Diagnostics.AddError("update user failed", err.Error())
		return
	}
	planParams := m.Params
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	m.Params = planParams
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.deleteUser(ctx, m.Domain.ValueString(), m.Number.ValueString())
	if err != nil && !errors.Is(err, ErrNotFound) {
		resp.Diagnostics.AddError("delete user failed", err.Error())
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("invalid import id", fmt.Sprintf("expected \"domain/number\", got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("number"), parts[1])...)
}
