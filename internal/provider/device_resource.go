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
	_ resource.Resource                = &deviceResource{}
	_ resource.ResourceWithImportState = &deviceResource{}
)

type deviceResource struct{ client *Client }

func NewDeviceResource() resource.Resource { return &deviceResource{} }

type deviceModel struct {
	ID          types.String `tfsdk:"id"`
	MAC         types.String `tfsdk:"mac"`
	Vendor      types.String `tfsdk:"vendor"`
	Model       types.String `tfsdk:"model"`
	Number      types.String `tfsdk:"number"`
	Domain      types.String `tfsdk:"domain"`
	DisplayName types.String `tfsdk:"display_name"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *deviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (r *deviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A provisioned physical SIP phone. The control-plane serves its config at " +
			"`GET /provision/<mac>` (the SIP password is read from the matching `freeswitch_user`).",
		Attributes: map[string]schema.Attribute{
			"id":     schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"mac":    schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Device MAC (any separators; normalized to lowercase, no separators)."},
			"vendor": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("yealink"), MarkdownDescription: "`yealink` | `grandstream` | `generic`."},
			"model":  schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
			"number": schema.StringAttribute{Required: true, MarkdownDescription: "SIP extension; its password is taken from the matching freeswitch_user."},
			"domain": schema.StringAttribute{Required: true},
			"display_name": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
			"enabled":      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"created_at":   schema.StringAttribute{Computed: true},
			"updated_at":   schema.StringAttribute{Computed: true},
		},
	}
}

func (r *deviceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *deviceResource) apiFromModel(m deviceModel) apiDevice {
	return apiDevice{
		MAC:         m.MAC.ValueString(),
		Vendor:      m.Vendor.ValueString(),
		Model:       m.Model.ValueString(),
		Number:      m.Number.ValueString(),
		Domain:      m.Domain.ValueString(),
		DisplayName: m.DisplayName.ValueString(),
		Enabled:     m.Enabled.ValueBool(),
	}
}

func (r *deviceResource) modelFromAPI(d *apiDevice, m *deviceModel) {
	m.ID = types.StringValue(d.ID)
	// Keep the config-supplied MAC verbatim: the server normalizes it (lowercase,
	// no separators) but lookups accept any form, so preserving the config value
	// avoids an "inconsistent result after apply" diff on a Required attribute.
	if m.MAC.IsNull() || m.MAC.IsUnknown() {
		m.MAC = types.StringValue(d.MAC)
	}
	m.Vendor = types.StringValue(d.Vendor)
	m.Model = types.StringValue(d.Model)
	m.Number = types.StringValue(d.Number)
	m.Domain = types.StringValue(d.Domain)
	m.DisplayName = types.StringValue(d.DisplayName)
	m.Enabled = types.BoolValue(d.Enabled)
	m.CreatedAt = types.StringValue(d.CreatedAt)
	m.UpdatedAt = types.StringValue(d.UpdatedAt)
}

func (r *deviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m deviceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.createDevice(ctx, r.apiFromModel(m))
	if err != nil {
		resp.Diagnostics.AddError("create device failed", err.Error())
		return
	}
	r.modelFromAPI(out, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *deviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m deviceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.getDevice(ctx, m.MAC.ValueString())
	if errors.Is(err, ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("read device failed", err.Error())
		return
	}
	r.modelFromAPI(out, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *deviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m deviceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.updateDevice(ctx, m.MAC.ValueString(), r.apiFromModel(m))
	if err != nil {
		resp.Diagnostics.AddError("update device failed", err.Error())
		return
	}
	r.modelFromAPI(out, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *deviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m deviceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.deleteDevice(ctx, m.MAC.ValueString()); err != nil && !errors.Is(err, ErrNotFound) {
		resp.Diagnostics.AddError("delete device failed", err.Error())
	}
}

func (r *deviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("mac"), req, resp)
}
