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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &dialplanResource{}
	_ resource.ResourceWithImportState = &dialplanResource{}
)

type dialplanResource struct{ client *Client }

func NewDialplanExtensionResource() resource.Resource { return &dialplanResource{} }

type dpActionModel struct {
	Application types.String `tfsdk:"application"`
	Data        types.String `tfsdk:"data"`
}

type dpConditionModel struct {
	Field      types.String    `tfsdk:"field"`
	Expression types.String    `tfsdk:"expression"`
	Time       types.Map       `tfsdk:"time"`
	Actions    []dpActionModel `tfsdk:"action"`
}

type dialplanModel struct {
	ID         types.String       `tfsdk:"id"`
	Name       types.String       `tfsdk:"name"`
	Domain     types.String       `tfsdk:"domain"`
	Context    types.String       `tfsdk:"context"`
	Priority   types.Int64        `tfsdk:"priority"`
	Enabled    types.Bool         `tfsdk:"enabled"`
	Conditions []dpConditionModel `tfsdk:"condition"`
	CreatedAt  types.String       `tfsdk:"created_at"`
	UpdatedAt  types.String       `tfsdk:"updated_at"`
}

func (r *dialplanResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dialplan_extension"
}

func (r *dialplanResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A FreeSWITCH dialplan extension (ordered conditions with actions).",
		Attributes: map[string]schema.Attribute{
			"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":     schema.StringAttribute{Required: true},
			"domain":   schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"context":  schema.StringAttribute{Required: true},
			"priority": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(100)},
			"enabled":  schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
		},
		Blocks: map[string]schema.Block{
			"condition": schema.ListNestedBlock{
				MarkdownDescription: "Ordered match conditions. At least one is required.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"field":      schema.StringAttribute{Optional: true, MarkdownDescription: "Channel variable to match (e.g. `destination_number`). Omit for a pure time gate."},
						"expression": schema.StringAttribute{Optional: true, MarkdownDescription: "Regex for `field`. Set together with `field`."},
						"time": schema.MapAttribute{
							Optional: true, ElementType: types.StringType,
							MarkdownDescription: "FreeSWITCH time-of-day attributes, e.g. `{ wday = \"2-6\", hour = \"9-17\" }`. Supported keys: wday, mday, mon, mweek, week, hour, minute, time-of-day, date-time.",
						},
					},
					Blocks: map[string]schema.Block{
						"action": schema.ListNestedBlock{
							MarkdownDescription: "Ordered actions for this condition. At least one is required.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"application": schema.StringAttribute{Required: true},
									"data":        schema.StringAttribute{Optional: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *dialplanResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *dialplanResource) apiFromModel(ctx context.Context, m dialplanModel) apiExtension {
	conds := make([]apiCondition, 0, len(m.Conditions))
	for _, c := range m.Conditions {
		actions := make([]apiAction, 0, len(c.Actions))
		for _, a := range c.Actions {
			actions = append(actions, apiAction{
				Application: a.Application.ValueString(),
				Data:        a.Data.ValueString(),
			})
		}
		timeAttrs, _ := tfMapToGo(ctx, c.Time)
		if len(timeAttrs) == 0 {
			timeAttrs = nil // omitempty on the wire
		}
		conds = append(conds, apiCondition{
			Field:      c.Field.ValueString(),
			Expression: c.Expression.ValueString(),
			Time:       timeAttrs,
			Actions:    actions,
		})
	}
	return apiExtension{
		Name:       m.Name.ValueString(),
		Domain:     m.Domain.ValueString(),
		Context:    m.Context.ValueString(),
		Priority:   m.Priority.ValueInt64(),
		Enabled:    m.Enabled.ValueBool(),
		Conditions: conds,
	}
}

func (r *dialplanResource) modelFromAPI(ctx context.Context, e *apiExtension, m *dialplanModel) {
	m.ID = types.StringValue(e.ID)
	m.Name = types.StringValue(e.Name)
	m.Domain = types.StringValue(e.Domain)
	m.Context = types.StringValue(e.Context)
	m.Priority = types.Int64Value(e.Priority)
	m.Enabled = types.BoolValue(e.Enabled)
	conds := make([]dpConditionModel, 0, len(e.Conditions))
	for _, c := range e.Conditions {
		actions := make([]dpActionModel, 0, len(c.Actions))
		for _, a := range c.Actions {
			am := dpActionModel{Application: types.StringValue(a.Application)}
			if a.Data == "" {
				am.Data = types.StringNull()
			} else {
				am.Data = types.StringValue(a.Data)
			}
			actions = append(actions, am)
		}
		cm := dpConditionModel{Actions: actions}
		cm.Field = optStr(c.Field)
		cm.Expression = optStr(c.Expression)
		if len(c.Time) == 0 {
			cm.Time = types.MapNull(types.StringType)
		} else {
			tm, _ := goMapToTF(ctx, c.Time)
			cm.Time = tm
		}
		conds = append(conds, cm)
	}
	m.Conditions = conds
	m.CreatedAt = types.StringValue(e.CreatedAt)
	m.UpdatedAt = types.StringValue(e.UpdatedAt)
}

func (r *dialplanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m dialplanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.createExtension(ctx, r.apiFromModel(ctx, m))
	if err != nil {
		resp.Diagnostics.AddError("create dialplan extension failed", err.Error())
		return
	}
	r.modelFromAPI(ctx, out, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *dialplanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m dialplanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.getExtension(ctx, m.ID.ValueString())
	if errors.Is(err, ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("read dialplan extension failed", err.Error())
		return
	}
	r.modelFromAPI(ctx, out, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *dialplanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m dialplanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state dialplanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.updateExtension(ctx, state.ID.ValueString(), r.apiFromModel(ctx, m))
	if err != nil {
		resp.Diagnostics.AddError("update dialplan extension failed", err.Error())
		return
	}
	r.modelFromAPI(ctx, out, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *dialplanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m dialplanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.deleteExtension(ctx, m.ID.ValueString())
	if err != nil && !errors.Is(err, ErrNotFound) {
		resp.Diagnostics.AddError("delete dialplan extension failed", err.Error())
	}
}

func (r *dialplanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
