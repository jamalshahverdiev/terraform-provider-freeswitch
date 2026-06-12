package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &reloadxmlResource{}

type reloadxmlResource struct{ client *Client }

func NewReloadXMLResource() resource.Resource { return &reloadxmlResource{} }

type reloadxmlModel struct {
	ID       types.String `tfsdk:"id"`
	Triggers types.Map    `tfsdk:"triggers"`
	Result   types.String `tfsdk:"result"`
}

func (r *reloadxmlResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_reloadxml"
}

func (r *reloadxmlResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs FreeSWITCH `reloadxml` via the control-plane ESL whenever `triggers` change. " +
			"Use `triggers` + `depends_on` to apply config after other resources change.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"triggers": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Arbitrary key/value pairs; any change re-runs reloadxml.",
				PlanModifiers:       []planmodifier.Map{mapplanmodifier.RequiresReplace()},
			},
			"result": schema.StringAttribute{Computed: true, MarkdownDescription: "ESL reply from the last reloadxml."},
		},
	}
}

func (r *reloadxmlResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *reloadxmlResource) run(ctx context.Context, m *reloadxmlModel, diags interface{ AddError(string, string) }) {
	out, err := r.client.reloadXML(ctx)
	if err != nil {
		diags.AddError("reloadxml failed", err.Error())
		return
	}
	m.ID = types.StringValue("reloadxml")
	m.Result = types.StringValue(out.Message)
}

func (r *reloadxmlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m reloadxmlModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.run(ctx, &m, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// Read is a no-op: the resource is an action, there is nothing to refresh.
func (r *reloadxmlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m reloadxmlModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// Update only fires for non-replacing changes; triggers force replacement, so
// this just persists the planned values.
func (r *reloadxmlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var m reloadxmlModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *reloadxmlResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Nothing to delete; the action has no persistent server-side object.
}
