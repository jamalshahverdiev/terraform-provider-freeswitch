package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &domainResource{}
	_ resource.ResourceWithImportState = &domainResource{}
)

type domainResource struct{ client *Client }

func NewDomainResource() resource.Resource { return &domainResource{} }

type domainModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Variables   types.Map    `tfsdk:"variables"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *domainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *domainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A FreeSWITCH directory domain (tenant).",
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Domain name (e.g. the FreeSWITCH $${domain})."},
			"description": schema.StringAttribute{Optional: true},
			"enabled":     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"variables":   schema.MapAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Domain-level variables."},
			"created_at":  schema.StringAttribute{Computed: true},
			"updated_at":  schema.StringAttribute{Computed: true},
		},
	}
}

func (r *domainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *domainResource) apiFromModel(ctx context.Context, m domainModel) apiDomain {
	vars, _ := tfMapToGo(ctx, m.Variables)
	return apiDomain{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Enabled:     m.Enabled.ValueBool(),
		Variables:   vars,
	}
}

func (r *domainResource) modelFromAPI(ctx context.Context, d *apiDomain, m *domainModel) error {
	m.ID = types.StringValue(d.ID)
	m.Name = types.StringValue(d.Name)
	if d.Description == "" {
		m.Description = types.StringNull()
	} else {
		m.Description = types.StringValue(d.Description)
	}
	m.Enabled = types.BoolValue(d.Enabled)
	vars, diags := goMapToTFPreserveNull(ctx, d.Variables, m.Variables)
	if diags.HasError() {
		return errors.New("variables conversion failed")
	}
	m.Variables = vars
	m.CreatedAt = types.StringValue(d.CreatedAt)
	m.UpdatedAt = types.StringValue(d.UpdatedAt)
	return nil
}

func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m domainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.createDomain(ctx, r.apiFromModel(ctx, m))
	if err != nil {
		resp.Diagnostics.AddError("create domain failed", err.Error())
		return
	}
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m domainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.getDomain(ctx, m.Name.ValueString())
	if errors.Is(err, ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("read domain failed", err.Error())
		return
	}
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m domainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.updateDomain(ctx, m.Name.ValueString(), r.apiFromModel(ctx, m))
	if err != nil {
		resp.Diagnostics.AddError("update domain failed", err.Error())
		return
	}
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m domainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.deleteDomain(ctx, m.Name.ValueString())
	if err != nil && !errors.Is(err, ErrNotFound) {
		resp.Diagnostics.AddError("delete domain failed", err.Error())
	}
}

func (r *domainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
