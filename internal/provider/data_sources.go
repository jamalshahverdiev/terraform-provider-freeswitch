package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func dsClient(data any, diags interface{ AddError(string, string) }) *Client {
	return clientFromProviderData(data, diags)
}

// ---------- freeswitch_domain ----------

type domainDataSource struct{ client *Client }

func NewDomainDataSource() datasource.DataSource { return &domainDataSource{} }

func (d *domainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}
func (d *domainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = dsClient(req.ProviderData, &resp.Diagnostics)
	}
}
func (d *domainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up an existing FreeSWITCH domain by name.",
		Attributes: map[string]schema.Attribute{
			"name":        schema.StringAttribute{Required: true},
			"id":          schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"enabled":     schema.BoolAttribute{Computed: true},
			"variables":   schema.MapAttribute{Computed: true, ElementType: types.StringType},
			"created_at":  schema.StringAttribute{Computed: true},
			"updated_at":  schema.StringAttribute{Computed: true},
		},
	}
}
func (d *domainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m domainModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := d.client.getDomain(ctx, m.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("read domain failed", err.Error())
		return
	}
	m.ID = types.StringValue(out.ID)
	if out.Description == "" {
		m.Description = types.StringNull()
	} else {
		m.Description = types.StringValue(out.Description)
	}
	m.Enabled = types.BoolValue(out.Enabled)
	vars, diags := goMapToTF(ctx, out.Variables)
	resp.Diagnostics.Append(diags...)
	m.Variables = vars
	m.CreatedAt = types.StringValue(out.CreatedAt)
	m.UpdatedAt = types.StringValue(out.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// ---------- freeswitch_user ----------

type userDataSource struct{ client *Client }

func NewUserDataSource() datasource.DataSource { return &userDataSource{} }

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}
func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = dsClient(req.ProviderData, &resp.Diagnostics)
	}
}
func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up an existing SIP user by domain and number.",
		Attributes: map[string]schema.Attribute{
			"domain":     schema.StringAttribute{Required: true},
			"number":     schema.StringAttribute{Required: true},
			"id":        schema.StringAttribute{Computed: true},
			"enabled":   schema.BoolAttribute{Computed: true},
			"params":    schema.MapAttribute{Computed: true, Sensitive: true, ElementType: types.StringType},
			"variables": schema.MapAttribute{Computed: true, ElementType: types.StringType},
			"voicemail": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Typed voicemail mailbox (null if the user has none).",
				Attributes: map[string]schema.Attribute{
					"enabled":     schema.BoolAttribute{Computed: true},
					"password":    schema.StringAttribute{Computed: true, Sensitive: true},
					"email":       schema.StringAttribute{Computed: true},
					"attach_file": schema.BoolAttribute{Computed: true},
					"email_all":   schema.BoolAttribute{Computed: true},
				},
			},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}
func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m userModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := d.client.getUser(ctx, m.Domain.ValueString(), m.Number.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("read user failed", err.Error())
		return
	}
	m.ID = types.StringValue(out.Domain + "/" + out.Number)
	m.Enabled = types.BoolValue(out.Enabled)
	params, d1 := goMapToTF(ctx, out.Params)
	vars, d2 := goMapToTF(ctx, out.Variables)
	resp.Diagnostics.Append(d1...)
	resp.Diagnostics.Append(d2...)
	m.Params = params
	m.Variables = vars
	m.Voicemail = voicemailFromAPI(out.Voicemail)
	m.CreatedAt = types.StringValue(out.CreatedAt)
	m.UpdatedAt = types.StringValue(out.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// ---------- freeswitch_gateway ----------

type gatewayDataSource struct{ client *Client }

func NewGatewayDataSource() datasource.DataSource { return &gatewayDataSource{} }

func (d *gatewayDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway"
}
func (d *gatewayDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = dsClient(req.ProviderData, &resp.Diagnostics)
	}
}
func (d *gatewayDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up an existing gateway by profile and name.",
		Attributes: map[string]schema.Attribute{
			"profile":    schema.StringAttribute{Required: true},
			"name":       schema.StringAttribute{Required: true},
			"id":         schema.StringAttribute{Computed: true},
			"enabled":    schema.BoolAttribute{Computed: true},
			"username":   schema.StringAttribute{Computed: true},
			"password":   schema.StringAttribute{Computed: true, Sensitive: true},
			"realm":      schema.StringAttribute{Computed: true},
			"proxy":      schema.StringAttribute{Computed: true},
			"register":   schema.BoolAttribute{Computed: true},
			"params":     schema.MapAttribute{Computed: true, ElementType: types.StringType},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}
func (d *gatewayDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m gatewayModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := d.client.getGateway(ctx, m.Profile.ValueString(), m.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("read gateway failed", err.Error())
		return
	}
	m.ID = types.StringValue(out.Profile + "/" + out.Name)
	m.Enabled = types.BoolValue(out.Enabled)
	m.Username = optStr(out.Username)
	m.Password = optStr(out.Password)
	m.Realm = optStr(out.Realm)
	m.Proxy = types.StringValue(out.Proxy)
	m.Register = types.BoolValue(out.Register)
	params, diags := goMapToTF(ctx, out.Params)
	resp.Diagnostics.Append(diags...)
	m.Params = params
	m.CreatedAt = types.StringValue(out.CreatedAt)
	m.UpdatedAt = types.StringValue(out.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// ---------- freeswitch_dialplan_extension ----------

type dialplanDataSource struct{ client *Client }

func NewDialplanExtensionDataSource() datasource.DataSource { return &dialplanDataSource{} }

type dialplanDataModel struct {
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

func (d *dialplanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dialplan_extension"
}
func (d *dialplanDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = dsClient(req.ProviderData, &resp.Diagnostics)
	}
}
func (d *dialplanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up an existing dialplan extension by id.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Required: true},
			"name":       schema.StringAttribute{Computed: true},
			"domain":     schema.StringAttribute{Computed: true},
			"context":    schema.StringAttribute{Computed: true},
			"priority":   schema.Int64Attribute{Computed: true},
			"enabled":    schema.BoolAttribute{Computed: true},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
			"condition": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field":      schema.StringAttribute{Computed: true},
						"expression": schema.StringAttribute{Computed: true},
						"action": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"application": schema.StringAttribute{Computed: true},
									"data":        schema.StringAttribute{Computed: true},
								},
							},
						},
					},
				},
			},
		},
	}
}
func (d *dialplanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m dialplanDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := d.client.getExtension(ctx, m.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("read dialplan extension failed", err.Error())
		return
	}
	m.Name = types.StringValue(out.Name)
	m.Domain = types.StringValue(out.Domain)
	m.Context = types.StringValue(out.Context)
	m.Priority = types.Int64Value(out.Priority)
	m.Enabled = types.BoolValue(out.Enabled)
	conds := make([]dpConditionModel, 0, len(out.Conditions))
	for _, c := range out.Conditions {
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
		conds = append(conds, dpConditionModel{
			Field:      types.StringValue(c.Field),
			Expression: types.StringValue(c.Expression),
			Actions:    actions,
		})
	}
	m.Conditions = conds
	m.CreatedAt = types.StringValue(out.CreatedAt)
	m.UpdatedAt = types.StringValue(out.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// ---------- freeswitch_gateway_status (runtime) ----------

type gatewayStatusDataSource struct{ client *Client }

func NewGatewayStatusDataSource() datasource.DataSource { return &gatewayStatusDataSource{} }

type gatewayStatusModel struct {
	Profile    types.String `tfsdk:"profile"`
	Name       types.String `tfsdk:"name"`
	Status     types.String `tfsdk:"status"`
	State      types.String `tfsdk:"state"`
	Attributes types.Map    `tfsdk:"attributes"`
}

func (d *gatewayStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway_status"
}
func (d *gatewayStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = dsClient(req.ProviderData, &resp.Diagnostics)
	}
}
func (d *gatewayStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runtime status of a Sofia gateway (via ESL `sofia status gateway`).",
		Attributes: map[string]schema.Attribute{
			"profile":    schema.StringAttribute{Required: true},
			"name":       schema.StringAttribute{Required: true},
			"status":     schema.StringAttribute{Computed: true},
			"state":      schema.StringAttribute{Computed: true},
			"attributes": schema.MapAttribute{Computed: true, ElementType: types.StringType},
		},
	}
}
func (d *gatewayStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m gatewayStatusModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := d.client.gatewayStatus(ctx, m.Profile.ValueString(), m.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("gateway status failed", err.Error())
		return
	}
	m.Status = types.StringValue(out.Status)
	m.State = types.StringValue(out.State)
	attrs, diags := goMapToTF(ctx, out.Attributes)
	resp.Diagnostics.Append(diags...)
	m.Attributes = attrs
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// ---------- freeswitch_user_registration (runtime) ----------

type userRegistrationDataSource struct{ client *Client }

func NewUserRegistrationDataSource() datasource.DataSource { return &userRegistrationDataSource{} }

type userRegistrationModel struct {
	User        types.String `tfsdk:"user"`
	Domain      types.String `tfsdk:"domain"`
	Registered  types.Bool   `tfsdk:"registered"`
	Contact     types.String `tfsdk:"contact"`
	Agent       types.String `tfsdk:"agent"`
	NetworkIP   types.String `tfsdk:"network_ip"`
	NetworkPort types.String `tfsdk:"network_port"`
	Expires     types.String `tfsdk:"expires"`
}

func (d *userRegistrationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_registration"
}
func (d *userRegistrationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = dsClient(req.ProviderData, &resp.Diagnostics)
	}
}
func (d *userRegistrationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runtime SIP registration state for a user (via ESL `show registrations`).",
		Attributes: map[string]schema.Attribute{
			"user":         schema.StringAttribute{Required: true},
			"domain":       schema.StringAttribute{Required: true},
			"registered":   schema.BoolAttribute{Computed: true},
			"contact":      schema.StringAttribute{Computed: true},
			"agent":        schema.StringAttribute{Computed: true},
			"network_ip":   schema.StringAttribute{Computed: true},
			"network_port": schema.StringAttribute{Computed: true},
			"expires":      schema.StringAttribute{Computed: true},
		},
	}
}
func (d *userRegistrationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m userRegistrationModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := d.client.registration(ctx, m.Domain.ValueString(), m.User.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("registration lookup failed", err.Error())
		return
	}
	m.Registered = types.BoolValue(out.Registered)
	m.Contact = types.StringValue(out.Contact)
	m.Agent = types.StringValue(out.Agent)
	m.NetworkIP = types.StringValue(out.NetworkIP)
	m.NetworkPort = types.StringValue(out.NetworkPort)
	m.Expires = types.StringValue(out.Expires)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// ---------- freeswitch_conference_status (runtime) ----------

type conferenceStatusDataSource struct{ client *Client }

func NewConferenceStatusDataSource() datasource.DataSource { return &conferenceStatusDataSource{} }

type confMemberModel struct {
	ID             types.String `tfsdk:"id"`
	CallerIDName   types.String `tfsdk:"caller_id_name"`
	CallerIDNumber types.String `tfsdk:"caller_id_number"`
	CanHear        types.Bool   `tfsdk:"can_hear"`
	CanSpeak       types.Bool   `tfsdk:"can_speak"`
	HasVideo       types.Bool   `tfsdk:"has_video"`
	Talking        types.Bool   `tfsdk:"talking"`
}

type conferenceStatusModel struct {
	Name        types.String      `tfsdk:"name"`
	Running     types.Bool        `tfsdk:"running"`
	MemberCount types.Int64       `tfsdk:"member_count"`
	RunTime     types.Int64       `tfsdk:"run_time"`
	Members     []confMemberModel `tfsdk:"members"`
}

func (d *conferenceStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_conference_status"
}
func (d *conferenceStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = dsClient(req.ProviderData, &resp.Diagnostics)
	}
}
func (d *conferenceStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Live participants of a running conference (via ESL `conference xml_list`).",
		Attributes: map[string]schema.Attribute{
			"name":         schema.StringAttribute{Required: true},
			"running":      schema.BoolAttribute{Computed: true},
			"member_count": schema.Int64Attribute{Computed: true},
			"run_time":     schema.Int64Attribute{Computed: true},
			"members": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":               schema.StringAttribute{Computed: true},
						"caller_id_name":   schema.StringAttribute{Computed: true},
						"caller_id_number": schema.StringAttribute{Computed: true},
						"can_hear":         schema.BoolAttribute{Computed: true},
						"can_speak":        schema.BoolAttribute{Computed: true},
						"has_video":        schema.BoolAttribute{Computed: true},
						"talking":          schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}
func (d *conferenceStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m conferenceStatusModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := d.client.conferenceStatus(ctx, m.Name.ValueString())
	if errors.Is(err, ErrNotFound) {
		m.Running = types.BoolValue(false)
		m.MemberCount = types.Int64Value(0)
		m.RunTime = types.Int64Value(0)
		m.Members = []confMemberModel{}
		resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("conference status lookup failed", err.Error())
		return
	}
	m.Running = types.BoolValue(true)
	m.MemberCount = types.Int64Value(out.MemberCount)
	m.RunTime = types.Int64Value(out.RunTime)
	m.Members = []confMemberModel{}
	for _, mem := range out.Members {
		m.Members = append(m.Members, confMemberModel{
			ID:             types.StringValue(mem.ID),
			CallerIDName:   types.StringValue(mem.CallerIDName),
			CallerIDNumber: types.StringValue(mem.CallerIDNumber),
			CanHear:        types.BoolValue(mem.CanHear),
			CanSpeak:       types.BoolValue(mem.CanSpeak),
			HasVideo:       types.BoolValue(mem.HasVideo),
			Talking:        types.BoolValue(mem.Talking),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// ---------- freeswitch_callcenter_queue ----------

type ccQueueDataSource struct{ client *Client }

func NewCCQueueDataSource() datasource.DataSource { return &ccQueueDataSource{} }

func (d *ccQueueDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_callcenter_queue"
}
func (d *ccQueueDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = dsClient(req.ProviderData, &resp.Diagnostics)
	}
}
func (d *ccQueueDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up an existing mod_callcenter queue by name.",
		Attributes: map[string]schema.Attribute{
			"name":                    schema.StringAttribute{Required: true},
			"strategy":                schema.StringAttribute{Computed: true},
			"moh_sound":               schema.StringAttribute{Computed: true},
			"time_base_score":         schema.StringAttribute{Computed: true},
			"max_wait_time":           schema.Int64Attribute{Computed: true},
			"discard_abandoned_after": schema.Int64Attribute{Computed: true},
			"params":                  schema.MapAttribute{Computed: true, ElementType: types.StringType},
		},
	}
}

type ccQueueDSModel struct {
	Name                  types.String `tfsdk:"name"`
	Strategy              types.String `tfsdk:"strategy"`
	MohSound              types.String `tfsdk:"moh_sound"`
	TimeBaseScore         types.String `tfsdk:"time_base_score"`
	MaxWaitTime           types.Int64  `tfsdk:"max_wait_time"`
	DiscardAbandonedAfter types.Int64  `tfsdk:"discard_abandoned_after"`
	Params                types.Map    `tfsdk:"params"`
}

func (d *ccQueueDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m ccQueueDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := d.client.getCCQueue(ctx, m.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("callcenter queue lookup failed", err.Error())
		return
	}
	m.Strategy = types.StringValue(out.Strategy)
	m.MohSound = types.StringValue(out.MohSound)
	m.TimeBaseScore = types.StringValue(out.TimeBaseScore)
	m.MaxWaitTime = types.Int64Value(out.MaxWaitTime)
	m.DiscardAbandonedAfter = types.Int64Value(out.DiscardAbandonedAfter)
	params, diags := goMapToTF(ctx, out.Params)
	resp.Diagnostics.Append(diags...)
	m.Params = params
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// ---------- freeswitch_callcenter_agent ----------

type ccAgentDataSource struct{ client *Client }

func NewCCAgentDataSource() datasource.DataSource { return &ccAgentDataSource{} }

func (d *ccAgentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_callcenter_agent"
}
func (d *ccAgentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = dsClient(req.ProviderData, &resp.Diagnostics)
	}
}
func (d *ccAgentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up an existing mod_callcenter agent by name.",
		Attributes: map[string]schema.Attribute{
			"name":          schema.StringAttribute{Required: true},
			"type":          schema.StringAttribute{Computed: true},
			"contact":       schema.StringAttribute{Computed: true},
			"status":        schema.StringAttribute{Computed: true},
			"max_no_answer": schema.Int64Attribute{Computed: true},
			"wrap_up_time":  schema.Int64Attribute{Computed: true},
		},
	}
}

type ccAgentDSModel struct {
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Contact     types.String `tfsdk:"contact"`
	Status      types.String `tfsdk:"status"`
	MaxNoAnswer types.Int64  `tfsdk:"max_no_answer"`
	WrapUpTime  types.Int64  `tfsdk:"wrap_up_time"`
}

func (d *ccAgentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m ccAgentDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := d.client.getCCAgent(ctx, m.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("callcenter agent lookup failed", err.Error())
		return
	}
	m.Type = types.StringValue(out.Type)
	m.Contact = types.StringValue(out.Contact)
	m.Status = types.StringValue(out.Status)
	m.MaxNoAnswer = types.Int64Value(out.MaxNoAnswer)
	m.WrapUpTime = types.Int64Value(out.WrapUpTime)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// ---------- freeswitch_callcenter_tier ----------

type ccTierDataSource struct{ client *Client }

func NewCCTierDataSource() datasource.DataSource { return &ccTierDataSource{} }

func (d *ccTierDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_callcenter_tier"
}
func (d *ccTierDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = dsClient(req.ProviderData, &resp.Diagnostics)
	}
}
func (d *ccTierDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up an existing mod_callcenter tier (queue/agent binding).",
		Attributes: map[string]schema.Attribute{
			"queue":    schema.StringAttribute{Required: true},
			"agent":    schema.StringAttribute{Required: true},
			"level":    schema.Int64Attribute{Computed: true},
			"position": schema.Int64Attribute{Computed: true},
		},
	}
}

type ccTierDSModel struct {
	Queue    types.String `tfsdk:"queue"`
	Agent    types.String `tfsdk:"agent"`
	Level    types.Int64  `tfsdk:"level"`
	Position types.Int64  `tfsdk:"position"`
}

func (d *ccTierDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m ccTierDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := d.client.getCCTier(ctx, m.Queue.ValueString(), m.Agent.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("callcenter tier lookup failed", err.Error())
		return
	}
	m.Level = types.Int64Value(out.Level)
	m.Position = types.Int64Value(out.Position)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// ---------- freeswitch_conference_profile ----------

type confProfileDataSource struct{ client *Client }

func NewConfProfileDataSource() datasource.DataSource { return &confProfileDataSource{} }

func (d *confProfileDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_conference_profile"
}
func (d *confProfileDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = dsClient(req.ProviderData, &resp.Diagnostics)
	}
}
func (d *confProfileDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up an existing conference profile by name.",
		Attributes: map[string]schema.Attribute{
			"name":              schema.StringAttribute{Required: true},
			"rate":              schema.Int64Attribute{Computed: true},
			"video_mode":        schema.StringAttribute{Computed: true},
			"video_layout":      schema.StringAttribute{Computed: true},
			"video_canvas_size": schema.StringAttribute{Computed: true},
			"video_fps":         schema.Int64Attribute{Computed: true},
			"moh_sound":         schema.StringAttribute{Computed: true},
			"auto_record":       schema.StringAttribute{Computed: true},
		},
	}
}

type confProfileDSModel struct {
	Name            types.String `tfsdk:"name"`
	Rate            types.Int64  `tfsdk:"rate"`
	VideoMode       types.String `tfsdk:"video_mode"`
	VideoLayout     types.String `tfsdk:"video_layout"`
	VideoCanvasSize types.String `tfsdk:"video_canvas_size"`
	VideoFPS        types.Int64  `tfsdk:"video_fps"`
	MohSound        types.String `tfsdk:"moh_sound"`
	AutoRecord      types.String `tfsdk:"auto_record"`
}

func (d *confProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m confProfileDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := d.client.getConfProfile(ctx, m.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("conference profile lookup failed", err.Error())
		return
	}
	m.Rate = types.Int64Value(out.Rate)
	m.VideoMode = types.StringValue(out.VideoMode)
	m.VideoLayout = types.StringValue(out.VideoLayout)
	m.VideoCanvasSize = types.StringValue(out.VideoCanvasSize)
	m.VideoFPS = types.Int64Value(out.VideoFPS)
	m.MohSound = types.StringValue(out.MohSound)
	m.AutoRecord = types.StringValue(out.AutoRecord)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// ---------- freeswitch_conference_room ----------

type confRoomDataSource struct{ client *Client }

func NewConfRoomDataSource() datasource.DataSource { return &confRoomDataSource{} }

func (d *confRoomDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_conference_room"
}
func (d *confRoomDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = dsClient(req.ProviderData, &resp.Diagnostics)
	}
}
func (d *confRoomDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up an existing conference room by name.",
		Attributes: map[string]schema.Attribute{
			"name":        schema.StringAttribute{Required: true},
			"number":      schema.StringAttribute{Computed: true},
			"domain":      schema.StringAttribute{Computed: true},
			"context":     schema.StringAttribute{Computed: true},
			"profile":     schema.StringAttribute{Computed: true},
			"max_members": schema.Int64Attribute{Computed: true},
			"enabled":     schema.BoolAttribute{Computed: true},
		},
	}
}

type confRoomDSModel struct {
	Name       types.String `tfsdk:"name"`
	Number     types.String `tfsdk:"number"`
	Domain     types.String `tfsdk:"domain"`
	Context    types.String `tfsdk:"context"`
	Profile    types.String `tfsdk:"profile"`
	MaxMembers types.Int64  `tfsdk:"max_members"`
	Enabled    types.Bool   `tfsdk:"enabled"`
}

func (d *confRoomDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m confRoomDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := d.client.getConfRoom(ctx, m.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("conference room lookup failed", err.Error())
		return
	}
	m.Number = types.StringValue(out.Number)
	m.Domain = types.StringValue(out.Domain)
	m.Context = types.StringValue(out.Context)
	m.Profile = types.StringValue(out.Profile)
	m.MaxMembers = types.Int64Value(out.MaxMembers)
	m.Enabled = types.BoolValue(out.Enabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// ---------- freeswitch_device ----------

type deviceDataSource struct{ client *Client }

func NewDeviceDataSource() datasource.DataSource { return &deviceDataSource{} }

func (d *deviceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}
func (d *deviceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = dsClient(req.ProviderData, &resp.Diagnostics)
	}
}
func (d *deviceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a provisioned phone by its MAC (any separators; normalized server-side).",
		Attributes: map[string]schema.Attribute{
			"mac":          schema.StringAttribute{Required: true},
			"id":           schema.StringAttribute{Computed: true},
			"vendor":       schema.StringAttribute{Computed: true},
			"model":        schema.StringAttribute{Computed: true},
			"number":       schema.StringAttribute{Computed: true},
			"domain":       schema.StringAttribute{Computed: true},
			"display_name": schema.StringAttribute{Computed: true},
			"enabled":      schema.BoolAttribute{Computed: true},
			"created_at":   schema.StringAttribute{Computed: true},
			"updated_at":   schema.StringAttribute{Computed: true},
		},
	}
}

type deviceDSModel struct {
	MAC         types.String `tfsdk:"mac"`
	ID          types.String `tfsdk:"id"`
	Vendor      types.String `tfsdk:"vendor"`
	Model       types.String `tfsdk:"model"`
	Number      types.String `tfsdk:"number"`
	Domain      types.String `tfsdk:"domain"`
	DisplayName types.String `tfsdk:"display_name"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (d *deviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m deviceDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := d.client.getDevice(ctx, m.MAC.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("device lookup failed", err.Error())
		return
	}
	m.ID = types.StringValue(out.ID)
	m.Vendor = types.StringValue(out.Vendor)
	m.Model = types.StringValue(out.Model)
	m.Number = types.StringValue(out.Number)
	m.Domain = types.StringValue(out.Domain)
	m.DisplayName = types.StringValue(out.DisplayName)
	m.Enabled = types.BoolValue(out.Enabled)
	m.CreatedAt = types.StringValue(out.CreatedAt)
	m.UpdatedAt = types.StringValue(out.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// ---------- freeswitch_voicemail ----------

type voicemailDataSource struct{ client *Client }

func NewVoicemailDataSource() datasource.DataSource { return &voicemailDataSource{} }

func (d *voicemailDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_voicemail"
}
func (d *voicemailDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = dsClient(req.ProviderData, &resp.Diagnostics)
	}
}
func (d *voicemailDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read a user's voicemail mailbox (messages + MWI counters) from freeswitch_core. Read-only.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{Required: true},
			"number": schema.StringAttribute{Required: true},
			"total":  schema.Int64Attribute{Computed: true},
			"unread": schema.Int64Attribute{Computed: true, MarkdownDescription: "Unread (MWI) message count."},
			"messages": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid":          schema.StringAttribute{Computed: true},
						"folder":        schema.StringAttribute{Computed: true},
						"cid_name":      schema.StringAttribute{Computed: true},
						"cid_number":    schema.StringAttribute{Computed: true},
						"created_epoch": schema.Int64Attribute{Computed: true},
						"read_epoch":    schema.Int64Attribute{Computed: true},
						"message_len":   schema.Int64Attribute{Computed: true},
						"read":          schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

type voicemailDSModel struct {
	Domain   types.String   `tfsdk:"domain"`
	Number   types.String   `tfsdk:"number"`
	Total    types.Int64    `tfsdk:"total"`
	Unread   types.Int64    `tfsdk:"unread"`
	Messages []vmMsgDSModel `tfsdk:"messages"`
}

type vmMsgDSModel struct {
	UUID         types.String `tfsdk:"uuid"`
	Folder       types.String `tfsdk:"folder"`
	CIDName      types.String `tfsdk:"cid_name"`
	CIDNumber    types.String `tfsdk:"cid_number"`
	CreatedEpoch types.Int64  `tfsdk:"created_epoch"`
	ReadEpoch    types.Int64  `tfsdk:"read_epoch"`
	MessageLen   types.Int64  `tfsdk:"message_len"`
	Read         types.Bool   `tfsdk:"read"`
}

func (d *voicemailDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m voicemailDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := d.client.getVoicemail(ctx, m.Domain.ValueString(), m.Number.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("read voicemail failed", err.Error())
		return
	}
	m.Total = types.Int64Value(out.Total)
	m.Unread = types.Int64Value(out.Unread)
	m.Messages = make([]vmMsgDSModel, 0, len(out.Messages))
	for _, msg := range out.Messages {
		m.Messages = append(m.Messages, vmMsgDSModel{
			UUID:         types.StringValue(msg.UUID),
			Folder:       types.StringValue(msg.Folder),
			CIDName:      types.StringValue(msg.CIDName),
			CIDNumber:    types.StringValue(msg.CIDNumber),
			CreatedEpoch: types.Int64Value(msg.CreatedEpoch),
			ReadEpoch:    types.Int64Value(msg.ReadEpoch),
			MessageLen:   types.Int64Value(msg.MessageLen),
			Read:         types.BoolValue(msg.Read),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}
