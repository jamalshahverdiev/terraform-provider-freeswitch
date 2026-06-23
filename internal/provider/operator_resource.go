package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &operatorResource{}
	_ resource.ResourceWithImportState = &operatorResource{}
)

type operatorResource struct{ client *Client }

func NewOperatorResource() resource.Resource { return &operatorResource{} }

type operatorModel struct {
	ID          types.String `tfsdk:"id"`
	Subject     types.String `tfsdk:"subject"`
	Domain      types.String `tfsdk:"domain"`
	Number      types.String `tfsdk:"number"`
	DisplayName types.String `tfsdk:"display_name"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *operatorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_operator"
}

func (r *operatorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Binds a Keycloak identity (subject) to a SIP extension. The webphone BFF " +
			"resolves the logged-in user's extension via this mapping. RBAC roles live in Keycloak, not here.",
		Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"subject": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Keycloak `sub` (or username); unique."},
			"domain":  schema.StringAttribute{Required: true},
			"number":  schema.StringAttribute{Required: true, MarkdownDescription: "SIP extension this identity registers as."},
			"display_name": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
			"enabled":      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"created_at":   schema.StringAttribute{Computed: true},
			"updated_at":   schema.StringAttribute{Computed: true},
		},
	}
}

func (r *operatorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *operatorResource) apiFromModel(m operatorModel) apiOperator {
	return apiOperator{
		Subject:     m.Subject.ValueString(),
		Domain:      m.Domain.ValueString(),
		Number:      m.Number.ValueString(),
		DisplayName: m.DisplayName.ValueString(),
		Enabled:     m.Enabled.ValueBool(),
	}
}

func (r *operatorResource) modelFromAPI(o *apiOperator, m *operatorModel) {
	m.ID = types.StringValue(o.ID)
	m.Subject = types.StringValue(o.Subject)
	m.Domain = types.StringValue(o.Domain)
	m.Number = types.StringValue(o.Number)
	m.DisplayName = types.StringValue(o.DisplayName)
	m.Enabled = types.BoolValue(o.Enabled)
	m.CreatedAt = types.StringValue(o.CreatedAt)
	m.UpdatedAt = types.StringValue(o.UpdatedAt)
}

func (r *operatorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m operatorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.createOperator(ctx, r.apiFromModel(m))
	if err != nil {
		resp.Diagnostics.AddError("create operator failed", err.Error())
		return
	}
	r.modelFromAPI(out, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *operatorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m operatorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.getOperator(ctx, m.Subject.ValueString())
	if errors.Is(err, ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("read operator failed", err.Error())
		return
	}
	r.modelFromAPI(out, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *operatorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m operatorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.updateOperator(ctx, m.Subject.ValueString(), r.apiFromModel(m))
	if err != nil {
		resp.Diagnostics.AddError("update operator failed", err.Error())
		return
	}
	r.modelFromAPI(out, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *operatorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m operatorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.deleteOperator(ctx, m.Subject.ValueString()); err != nil && !errors.Is(err, ErrNotFound) {
		resp.Diagnostics.AddError("delete operator failed", err.Error())
	}
}

func (r *operatorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("subject"), req, resp)
}
