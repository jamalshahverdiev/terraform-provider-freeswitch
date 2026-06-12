package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ===================== freeswitch_conference_profile =====================

var (
	_ resource.Resource                = &confProfileResource{}
	_ resource.ResourceWithImportState = &confProfileResource{}
)

type confProfileResource struct{ client *Client }

func NewConfProfileResource() resource.Resource { return &confProfileResource{} }

type confProfileModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Rate            types.Int64  `tfsdk:"rate"`
	IntervalMs      types.Int64  `tfsdk:"interval_ms"`
	EnergyLevel     types.Int64  `tfsdk:"energy_level"`
	ComfortNoise    types.Bool   `tfsdk:"comfort_noise"`
	MohSound        types.String `tfsdk:"moh_sound"`
	VideoMode       types.String `tfsdk:"video_mode"`
	VideoLayout     types.String `tfsdk:"video_layout"`
	VideoCanvasSize types.String `tfsdk:"video_canvas_size"`
	VideoFPS        types.Int64  `tfsdk:"video_fps"`
	AutoRecord      types.String `tfsdk:"auto_record"`
	Params          types.Map    `tfsdk:"params"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func (r *confProfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_conference_profile"
}

func (r *confProfileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A mod_conference profile (settings group). Set `video_mode = \"mux\"` for a video conference with composed layouts. Profiles are read when a NEW conference starts — no reload needed.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":          schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"rate":          schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(48000)},
			"interval_ms":   schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(20)},
			"energy_level":  schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(200)},
			"comfort_noise": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"moh_sound":     schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("local_stream://moh")},
			"video_mode":    schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Empty = audio only; `mux` = composed video conference."},
			"video_layout":  schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("group:grid")},
			"video_canvas_size": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("1280x720")},
			"video_fps":         schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(15)},
			"auto_record":       schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Recording path template; empty disables recording."},
			"params":            schema.MapAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Extra profile params merged into the rendered XML."},
			"created_at":        schema.StringAttribute{Computed: true},
			"updated_at":        schema.StringAttribute{Computed: true},
		},
	}
}

func (r *confProfileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *confProfileResource) apiFromModel(ctx context.Context, m confProfileModel) apiConfProfile {
	params, _ := tfMapToGo(ctx, m.Params)
	return apiConfProfile{
		Name:            m.Name.ValueString(),
		Rate:            m.Rate.ValueInt64(),
		IntervalMs:      m.IntervalMs.ValueInt64(),
		EnergyLevel:     m.EnergyLevel.ValueInt64(),
		ComfortNoise:    m.ComfortNoise.ValueBool(),
		MohSound:        m.MohSound.ValueString(),
		VideoMode:       m.VideoMode.ValueString(),
		VideoLayout:     m.VideoLayout.ValueString(),
		VideoCanvasSize: m.VideoCanvasSize.ValueString(),
		VideoFPS:        m.VideoFPS.ValueInt64(),
		AutoRecord:      m.AutoRecord.ValueString(),
		Params:          params,
	}
}

func (r *confProfileResource) modelFromAPI(ctx context.Context, p *apiConfProfile, m *confProfileModel) error {
	m.ID = types.StringValue(p.Name)
	m.Name = types.StringValue(p.Name)
	m.Rate = types.Int64Value(p.Rate)
	m.IntervalMs = types.Int64Value(p.IntervalMs)
	m.EnergyLevel = types.Int64Value(p.EnergyLevel)
	m.ComfortNoise = types.BoolValue(p.ComfortNoise)
	m.MohSound = types.StringValue(p.MohSound)
	m.VideoMode = types.StringValue(p.VideoMode)
	m.VideoLayout = types.StringValue(p.VideoLayout)
	m.VideoCanvasSize = types.StringValue(p.VideoCanvasSize)
	m.VideoFPS = types.Int64Value(p.VideoFPS)
	m.AutoRecord = types.StringValue(p.AutoRecord)
	params, diags := goMapToTFPreserveNull(ctx, p.Params, m.Params)
	if diags.HasError() {
		return errors.New("params conversion failed")
	}
	m.Params = params
	m.CreatedAt = types.StringValue(p.CreatedAt)
	m.UpdatedAt = types.StringValue(p.UpdatedAt)
	return nil
}

func (r *confProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m confProfileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.createConfProfile(ctx, r.apiFromModel(ctx, m))
	if err != nil {
		resp.Diagnostics.AddError("create conference profile failed", err.Error())
		return
	}
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *confProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m confProfileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.getConfProfile(ctx, m.Name.ValueString())
	if errors.Is(err, ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("read conference profile failed", err.Error())
		return
	}
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *confProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m confProfileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.updateConfProfile(ctx, m.Name.ValueString(), r.apiFromModel(ctx, m))
	if err != nil {
		resp.Diagnostics.AddError("update conference profile failed", err.Error())
		return
	}
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *confProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m confProfileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.deleteConfProfile(ctx, m.Name.ValueString()); err != nil && !errors.Is(err, ErrNotFound) {
		resp.Diagnostics.AddError("delete conference profile failed", err.Error())
	}
}

func (r *confProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// ===================== freeswitch_conference_room =====================

var (
	_ resource.Resource                = &confRoomResource{}
	_ resource.ResourceWithImportState = &confRoomResource{}
)

type confRoomResource struct{ client *Client }

func NewConfRoomResource() resource.Resource { return &confRoomResource{} }

type confRoomModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Number     types.String `tfsdk:"number"`
	Domain     types.String `tfsdk:"domain"`
	Context    types.String `tfsdk:"context"`
	Profile    types.String `tfsdk:"profile"`
	Pin        types.String `tfsdk:"pin"`
	MaxMembers types.Int64  `tfsdk:"max_members"`
	Priority   types.Int64  `tfsdk:"priority"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

func (r *confRoomResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_conference_room"
}

func (r *confRoomResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A conference room: the room itself plus the dialplan extension callers dial to enter it (rendered into `/xml/dialplan` automatically).",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":        schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Conference name, e.g. `standup`."},
			"number":      schema.StringAttribute{Required: true, MarkdownDescription: "Dialable number, e.g. `3500`."},
			"domain":      schema.StringAttribute{Required: true},
			"context":     schema.StringAttribute{Required: true, MarkdownDescription: "Dialplan context the entry extension is rendered into."},
			"profile":     schema.StringAttribute{Required: true, MarkdownDescription: "Name of a `freeswitch_conference_profile`."},
			"pin":         schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Sensitive: true},
			"max_members": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(0), MarkdownDescription: "0 = unlimited."},
			"priority":    schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(5), MarkdownDescription: "Dialplan ordering — keep lower than broad catch-all extensions."},
			"enabled":     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"created_at":  schema.StringAttribute{Computed: true},
			"updated_at":  schema.StringAttribute{Computed: true},
		},
	}
}

func (r *confRoomResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *confRoomResource) apiFromModel(m confRoomModel) apiConfRoom {
	return apiConfRoom{
		Name:       m.Name.ValueString(),
		Number:     m.Number.ValueString(),
		Domain:     m.Domain.ValueString(),
		Context:    m.Context.ValueString(),
		Profile:    m.Profile.ValueString(),
		Pin:        m.Pin.ValueString(),
		MaxMembers: m.MaxMembers.ValueInt64(),
		Priority:   m.Priority.ValueInt64(),
		Enabled:    m.Enabled.ValueBool(),
	}
}

func (r *confRoomResource) modelFromAPI(room *apiConfRoom, m *confRoomModel) {
	m.ID = types.StringValue(room.Name)
	m.Name = types.StringValue(room.Name)
	m.Number = types.StringValue(room.Number)
	m.Domain = types.StringValue(room.Domain)
	m.Context = types.StringValue(room.Context)
	m.Profile = types.StringValue(room.Profile)
	m.Pin = types.StringValue(room.Pin)
	m.MaxMembers = types.Int64Value(room.MaxMembers)
	m.Priority = types.Int64Value(room.Priority)
	m.Enabled = types.BoolValue(room.Enabled)
	m.CreatedAt = types.StringValue(room.CreatedAt)
	m.UpdatedAt = types.StringValue(room.UpdatedAt)
}

func (r *confRoomResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m confRoomModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.createConfRoom(ctx, r.apiFromModel(m))
	if err != nil {
		resp.Diagnostics.AddError("create conference room failed", err.Error())
		return
	}
	r.modelFromAPI(out, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *confRoomResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m confRoomModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.getConfRoom(ctx, m.Name.ValueString())
	if errors.Is(err, ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("read conference room failed", err.Error())
		return
	}
	r.modelFromAPI(out, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *confRoomResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m confRoomModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.updateConfRoom(ctx, m.Name.ValueString(), r.apiFromModel(m))
	if err != nil {
		resp.Diagnostics.AddError("update conference room failed", err.Error())
		return
	}
	r.modelFromAPI(out, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *confRoomResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m confRoomModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.deleteConfRoom(ctx, m.Name.ValueString()); err != nil && !errors.Is(err, ErrNotFound) {
		resp.Diagnostics.AddError("delete conference room failed", err.Error())
	}
}

func (r *confRoomResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
