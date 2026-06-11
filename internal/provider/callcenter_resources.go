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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ===================== freeswitch_callcenter_queue =====================

var (
	_ resource.Resource                = &ccQueueResource{}
	_ resource.ResourceWithImportState = &ccQueueResource{}
)

type ccQueueResource struct{ client *Client }

func NewCCQueueResource() resource.Resource { return &ccQueueResource{} }

type ccQueueModel struct {
	ID                                types.String `tfsdk:"id"`
	Name                              types.String `tfsdk:"name"`
	Strategy                          types.String `tfsdk:"strategy"`
	MohSound                          types.String `tfsdk:"moh_sound"`
	TimeBaseScore                     types.String `tfsdk:"time_base_score"`
	MaxWaitTime                       types.Int64  `tfsdk:"max_wait_time"`
	MaxWaitTimeWithNoAgent            types.Int64  `tfsdk:"max_wait_time_with_no_agent"`
	MaxWaitTimeWithNoAgentTimeReached types.Int64  `tfsdk:"max_wait_time_with_no_agent_time_reached"`
	TierRulesApply                    types.Bool   `tfsdk:"tier_rules_apply"`
	TierRuleWaitSecond                types.Int64  `tfsdk:"tier_rule_wait_second"`
	TierRuleWaitMultiplyLevel         types.Bool   `tfsdk:"tier_rule_wait_multiply_level"`
	TierRuleNoAgentNoWait             types.Bool   `tfsdk:"tier_rule_no_agent_no_wait"`
	DiscardAbandonedAfter             types.Int64  `tfsdk:"discard_abandoned_after"`
	AbandonedResumeAllowed            types.Bool   `tfsdk:"abandoned_resume_allowed"`
	Params                            types.Map    `tfsdk:"params"`
	CreatedAt                         types.String `tfsdk:"created_at"`
	UpdatedAt                         types.String `tfsdk:"updated_at"`
}

func (r *ccQueueResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_callcenter_queue"
}

func (r *ccQueueResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A mod_callcenter queue (served to FreeSWITCH from the control-plane DB).",
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Queue name, e.g. `support@192.168.48.143`."},
			"strategy":        schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("longest-idle-agent")},
			"moh_sound":       schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("local_stream://moh")},
			"time_base_score": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("system")},
			"max_wait_time":   schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(0)},
			"max_wait_time_with_no_agent":              schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(0)},
			"max_wait_time_with_no_agent_time_reached": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(5)},
			"tier_rules_apply":              schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"tier_rule_wait_second":         schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(300)},
			"tier_rule_wait_multiply_level": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"tier_rule_no_agent_no_wait":    schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"discard_abandoned_after":       schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(60)},
			"abandoned_resume_allowed":      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"params":                        schema.MapAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Extra queue params merged into the rendered XML."},
			"created_at":                    schema.StringAttribute{Computed: true},
			"updated_at":                    schema.StringAttribute{Computed: true},
		},
	}
}

func (r *ccQueueResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *ccQueueResource) apiFromModel(ctx context.Context, m ccQueueModel) apiCCQueue {
	params, _ := tfMapToGo(ctx, m.Params)
	return apiCCQueue{
		Name:                              m.Name.ValueString(),
		Strategy:                          m.Strategy.ValueString(),
		MohSound:                          m.MohSound.ValueString(),
		TimeBaseScore:                     m.TimeBaseScore.ValueString(),
		MaxWaitTime:                       m.MaxWaitTime.ValueInt64(),
		MaxWaitTimeWithNoAgent:            m.MaxWaitTimeWithNoAgent.ValueInt64(),
		MaxWaitTimeWithNoAgentTimeReached: m.MaxWaitTimeWithNoAgentTimeReached.ValueInt64(),
		TierRulesApply:                    m.TierRulesApply.ValueBool(),
		TierRuleWaitSecond:                m.TierRuleWaitSecond.ValueInt64(),
		TierRuleWaitMultiplyLevel:         m.TierRuleWaitMultiplyLevel.ValueBool(),
		TierRuleNoAgentNoWait:             m.TierRuleNoAgentNoWait.ValueBool(),
		DiscardAbandonedAfter:             m.DiscardAbandonedAfter.ValueInt64(),
		AbandonedResumeAllowed:            m.AbandonedResumeAllowed.ValueBool(),
		Params:                            params,
	}
}

func (r *ccQueueResource) modelFromAPI(ctx context.Context, q *apiCCQueue, m *ccQueueModel) error {
	m.ID = types.StringValue(q.Name)
	m.Name = types.StringValue(q.Name)
	m.Strategy = types.StringValue(q.Strategy)
	m.MohSound = types.StringValue(q.MohSound)
	m.TimeBaseScore = types.StringValue(q.TimeBaseScore)
	m.MaxWaitTime = types.Int64Value(q.MaxWaitTime)
	m.MaxWaitTimeWithNoAgent = types.Int64Value(q.MaxWaitTimeWithNoAgent)
	m.MaxWaitTimeWithNoAgentTimeReached = types.Int64Value(q.MaxWaitTimeWithNoAgentTimeReached)
	m.TierRulesApply = types.BoolValue(q.TierRulesApply)
	m.TierRuleWaitSecond = types.Int64Value(q.TierRuleWaitSecond)
	m.TierRuleWaitMultiplyLevel = types.BoolValue(q.TierRuleWaitMultiplyLevel)
	m.TierRuleNoAgentNoWait = types.BoolValue(q.TierRuleNoAgentNoWait)
	m.DiscardAbandonedAfter = types.Int64Value(q.DiscardAbandonedAfter)
	m.AbandonedResumeAllowed = types.BoolValue(q.AbandonedResumeAllowed)
	params, diags := goMapToTFPreserveNull(ctx, q.Params, m.Params)
	if diags.HasError() {
		return errors.New("params conversion failed")
	}
	m.Params = params
	m.CreatedAt = types.StringValue(q.CreatedAt)
	m.UpdatedAt = types.StringValue(q.UpdatedAt)
	return nil
}

func (r *ccQueueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m ccQueueModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.createCCQueue(ctx, r.apiFromModel(ctx, m))
	if err != nil {
		resp.Diagnostics.AddError("create callcenter queue failed", err.Error())
		return
	}
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ccQueueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m ccQueueModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.getCCQueue(ctx, m.Name.ValueString())
	if errors.Is(err, ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("read callcenter queue failed", err.Error())
		return
	}
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ccQueueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m ccQueueModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.updateCCQueue(ctx, m.Name.ValueString(), r.apiFromModel(ctx, m))
	if err != nil {
		resp.Diagnostics.AddError("update callcenter queue failed", err.Error())
		return
	}
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ccQueueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m ccQueueModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.deleteCCQueue(ctx, m.Name.ValueString()); err != nil && !errors.Is(err, ErrNotFound) {
		resp.Diagnostics.AddError("delete callcenter queue failed", err.Error())
	}
}

func (r *ccQueueResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// ===================== freeswitch_callcenter_agent =====================

var (
	_ resource.Resource                = &ccAgentResource{}
	_ resource.ResourceWithImportState = &ccAgentResource{}
)

type ccAgentResource struct{ client *Client }

func NewCCAgentResource() resource.Resource { return &ccAgentResource{} }

type ccAgentModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Type              types.String `tfsdk:"type"`
	Contact           types.String `tfsdk:"contact"`
	Status            types.String `tfsdk:"status"`
	MaxNoAnswer       types.Int64  `tfsdk:"max_no_answer"`
	WrapUpTime        types.Int64  `tfsdk:"wrap_up_time"`
	RejectDelayTime   types.Int64  `tfsdk:"reject_delay_time"`
	BusyDelayTime     types.Int64  `tfsdk:"busy_delay_time"`
	NoAnswerDelayTime types.Int64  `tfsdk:"no_answer_delay_time"`
	Params            types.Map    `tfsdk:"params"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

func (r *ccAgentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_callcenter_agent"
}

func (r *ccAgentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A mod_callcenter agent. `contact` is usually `user/<number>@<domain>` of a freeswitch_user.",
		Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":    schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Agent name, e.g. `4201@192.168.48.143`."},
			"type":    schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("callback")},
			"contact": schema.StringAttribute{Required: true},
			"status":  schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("Available"), MarkdownDescription: "Initial status on load (live status is runtime state)."},
			"max_no_answer":        schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(3)},
			"wrap_up_time":         schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(10)},
			"reject_delay_time":    schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(3)},
			"busy_delay_time":      schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(60)},
			"no_answer_delay_time": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(60)},
			"params":               schema.MapAttribute{Optional: true, ElementType: types.StringType},
			"created_at":           schema.StringAttribute{Computed: true},
			"updated_at":           schema.StringAttribute{Computed: true},
		},
	}
}

func (r *ccAgentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *ccAgentResource) apiFromModel(ctx context.Context, m ccAgentModel) apiCCAgent {
	params, _ := tfMapToGo(ctx, m.Params)
	return apiCCAgent{
		Name:              m.Name.ValueString(),
		Type:              m.Type.ValueString(),
		Contact:           m.Contact.ValueString(),
		Status:            m.Status.ValueString(),
		MaxNoAnswer:       m.MaxNoAnswer.ValueInt64(),
		WrapUpTime:        m.WrapUpTime.ValueInt64(),
		RejectDelayTime:   m.RejectDelayTime.ValueInt64(),
		BusyDelayTime:     m.BusyDelayTime.ValueInt64(),
		NoAnswerDelayTime: m.NoAnswerDelayTime.ValueInt64(),
		Params:            params,
	}
}

func (r *ccAgentResource) modelFromAPI(ctx context.Context, a *apiCCAgent, m *ccAgentModel) error {
	m.ID = types.StringValue(a.Name)
	m.Name = types.StringValue(a.Name)
	m.Type = types.StringValue(a.Type)
	m.Contact = types.StringValue(a.Contact)
	m.Status = types.StringValue(a.Status)
	m.MaxNoAnswer = types.Int64Value(a.MaxNoAnswer)
	m.WrapUpTime = types.Int64Value(a.WrapUpTime)
	m.RejectDelayTime = types.Int64Value(a.RejectDelayTime)
	m.BusyDelayTime = types.Int64Value(a.BusyDelayTime)
	m.NoAnswerDelayTime = types.Int64Value(a.NoAnswerDelayTime)
	params, diags := goMapToTFPreserveNull(ctx, a.Params, m.Params)
	if diags.HasError() {
		return errors.New("params conversion failed")
	}
	m.Params = params
	m.CreatedAt = types.StringValue(a.CreatedAt)
	m.UpdatedAt = types.StringValue(a.UpdatedAt)
	return nil
}

func (r *ccAgentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m ccAgentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.createCCAgent(ctx, r.apiFromModel(ctx, m))
	if err != nil {
		resp.Diagnostics.AddError("create callcenter agent failed", err.Error())
		return
	}
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ccAgentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m ccAgentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.getCCAgent(ctx, m.Name.ValueString())
	if errors.Is(err, ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("read callcenter agent failed", err.Error())
		return
	}
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ccAgentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m ccAgentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.updateCCAgent(ctx, m.Name.ValueString(), r.apiFromModel(ctx, m))
	if err != nil {
		resp.Diagnostics.AddError("update callcenter agent failed", err.Error())
		return
	}
	if err := r.modelFromAPI(ctx, out, &m); err != nil {
		resp.Diagnostics.AddError("response mapping failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ccAgentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m ccAgentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.deleteCCAgent(ctx, m.Name.ValueString()); err != nil && !errors.Is(err, ErrNotFound) {
		resp.Diagnostics.AddError("delete callcenter agent failed", err.Error())
	}
}

func (r *ccAgentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// ===================== freeswitch_callcenter_tier =====================

var (
	_ resource.Resource                = &ccTierResource{}
	_ resource.ResourceWithImportState = &ccTierResource{}
)

type ccTierResource struct{ client *Client }

func NewCCTierResource() resource.Resource { return &ccTierResource{} }

type ccTierModel struct {
	ID        types.String `tfsdk:"id"`
	Queue     types.String `tfsdk:"queue"`
	Agent     types.String `tfsdk:"agent"`
	Level     types.Int64  `tfsdk:"level"`
	Position  types.Int64  `tfsdk:"position"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *ccTierResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_callcenter_tier"
}

func (r *ccTierResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Binds a callcenter agent to a queue (with level/position).",
		Attributes: map[string]schema.Attribute{
			"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"queue":    schema.StringAttribute{Required: true, PlanModifiers: replace},
			"agent":    schema.StringAttribute{Required: true, PlanModifiers: replace},
			"level":    schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(1)},
			"position": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(1)},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *ccTierResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *ccTierResource) modelFromAPI(t *apiCCTier, m *ccTierModel) {
	m.ID = types.StringValue(t.Queue + "/" + t.Agent)
	m.Queue = types.StringValue(t.Queue)
	m.Agent = types.StringValue(t.Agent)
	m.Level = types.Int64Value(t.Level)
	m.Position = types.Int64Value(t.Position)
	m.CreatedAt = types.StringValue(t.CreatedAt)
	m.UpdatedAt = types.StringValue(t.UpdatedAt)
}

func (r *ccTierResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m ccTierModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.createCCTier(ctx, apiCCTier{
		Queue: m.Queue.ValueString(), Agent: m.Agent.ValueString(),
		Level: m.Level.ValueInt64(), Position: m.Position.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("create callcenter tier failed", err.Error())
		return
	}
	r.modelFromAPI(out, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ccTierResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m ccTierModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.getCCTier(ctx, m.Queue.ValueString(), m.Agent.ValueString())
	if errors.Is(err, ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("read callcenter tier failed", err.Error())
		return
	}
	r.modelFromAPI(out, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ccTierResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m ccTierModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.updateCCTier(ctx, m.Queue.ValueString(), m.Agent.ValueString(), apiCCTier{
		Level: m.Level.ValueInt64(), Position: m.Position.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("update callcenter tier failed", err.Error())
		return
	}
	r.modelFromAPI(out, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ccTierResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m ccTierModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.deleteCCTier(ctx, m.Queue.ValueString(), m.Agent.ValueString()); err != nil && !errors.Is(err, ErrNotFound) {
		resp.Diagnostics.AddError("delete callcenter tier failed", err.Error())
	}
}

func (r *ccTierResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("invalid import id", fmt.Sprintf("expected \"queue/agent\", got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("queue"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("agent"), parts[1])...)
}

// ===================== freeswitch_callcenter_reload =====================

var _ resource.Resource = &ccReloadResource{}

type ccReloadResource struct{ client *Client }

func NewCCReloadResource() resource.Resource { return &ccReloadResource{} }

type ccReloadModel struct {
	ID       types.String `tfsdk:"id"`
	Triggers types.Map    `tfsdk:"triggers"`
	Result   types.String `tfsdk:"result"`
}

func (r *ccReloadResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_callcenter_reload"
}

func (r *ccReloadResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reloads mod_callcenter (re-reads the DB-served callcenter.conf) whenever `triggers` change.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"triggers": schema.MapAttribute{
				Optional:      true,
				ElementType:   types.StringType,
				PlanModifiers: []planmodifier.Map{mapplanmodifier.RequiresReplace()},
			},
			"result": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *ccReloadResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *ccReloadResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m ccReloadModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.ccReload(ctx)
	if err != nil {
		resp.Diagnostics.AddError("callcenter reload failed", err.Error())
		return
	}
	m.ID = types.StringValue("callcenter-reload")
	m.Result = types.StringValue(out.Message)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ccReloadResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m ccReloadModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ccReloadResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m ccReloadModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ccReloadResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {}
