package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kenzo0107/terraform-provider-langfuse/langfuse"
)

var _ resource.Resource = &scoreResource{}
var _ resource.ResourceWithImportState = &scoreResource{}

func newScoreResource() resource.Resource {
	return &scoreResource{}
}

type scoreResource struct {
	client *langfuse.Client
}

type scoreResourceModel struct {
	ID            types.String  `tfsdk:"id"`
	Name          types.String  `tfsdk:"name"`
	Value         types.Float64 `tfsdk:"value"`
	StringValue   types.String  `tfsdk:"string_value"`
	DataType      types.String  `tfsdk:"data_type"`
	TraceID       types.String  `tfsdk:"trace_id"`
	ObservationID types.String  `tfsdk:"observation_id"`
	ConfigID      types.String  `tfsdk:"config_id"`
	Comment       types.String  `tfsdk:"comment"`
	Environment   types.String  `tfsdk:"environment"`
	Source        types.String  `tfsdk:"source"`
}

func (r *scoreResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_score"
}

func (r *scoreResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Langfuse score. All attributes are immutable; any change requires replacement.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the score.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the score. Changing this creates a new score.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.Float64Attribute{
				MarkdownDescription: "The numeric value of the score (for `NUMERIC` or `BOOLEAN` data types). Changing this creates a new score.",
				Optional:            true,
			},
			"string_value": schema.StringAttribute{
				MarkdownDescription: "The string value of the score (for `CATEGORICAL`, `TEXT`, or `CORRECTION` data types). Changing this creates a new score.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data_type": schema.StringAttribute{
				MarkdownDescription: "The data type of the score. Must be `NUMERIC`, `BOOLEAN`, `CATEGORICAL`, `TEXT`, or `CORRECTION`. Changing this creates a new score.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"trace_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the trace to associate the score with. Changing this creates a new score.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"observation_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the observation to associate the score with. Changing this creates a new score.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the score config. Changing this creates a new score.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "An optional comment for the score. Changing this creates a new score.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "The environment associated with the score (read-only after creation).",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "The source of the score (read-only, set by Langfuse).",
				Computed:            true,
			},
		},
	}
}

func (r *scoreResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*langfuse.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *langfuse.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func applyScoreToModel(s *langfuse.Score, model *scoreResourceModel) {
	model.ID = types.StringValue(s.ID)
	model.Name = types.StringValue(s.Name)
	model.Source = types.StringValue(s.Source)
	model.Environment = types.StringValue(s.Environment)

	if s.Value != nil {
		model.Value = types.Float64Value(*s.Value)
	} else {
		model.Value = types.Float64Null()
	}
	if s.StringValue != nil {
		model.StringValue = types.StringValue(*s.StringValue)
	} else {
		model.StringValue = types.StringNull()
	}
	if s.DataType != nil {
		model.DataType = types.StringValue(*s.DataType)
	} else {
		model.DataType = types.StringNull()
	}
	if s.TraceID != nil {
		model.TraceID = types.StringValue(*s.TraceID)
	} else {
		model.TraceID = types.StringNull()
	}
	if s.ObservationID != nil {
		model.ObservationID = types.StringValue(*s.ObservationID)
	} else {
		model.ObservationID = types.StringNull()
	}
	if s.ConfigID != nil {
		model.ConfigID = types.StringValue(*s.ConfigID)
	} else {
		model.ConfigID = types.StringNull()
	}
	if s.Comment != nil {
		model.Comment = types.StringValue(*s.Comment)
	} else {
		model.Comment = types.StringNull()
	}
}

func (r *scoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan scoreResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &langfuse.CreateScoreRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Value.IsNull() && !plan.Value.IsUnknown() {
		v := plan.Value.ValueFloat64()
		createReq.Value = &v
	}
	if !plan.StringValue.IsNull() && !plan.StringValue.IsUnknown() {
		v := plan.StringValue.ValueString()
		createReq.StringValue = &v
	}
	if !plan.DataType.IsNull() && !plan.DataType.IsUnknown() {
		v := plan.DataType.ValueString()
		createReq.DataType = &v
	}
	if !plan.TraceID.IsNull() && !plan.TraceID.IsUnknown() {
		v := plan.TraceID.ValueString()
		createReq.TraceID = &v
	}
	if !plan.ObservationID.IsNull() && !plan.ObservationID.IsUnknown() {
		v := plan.ObservationID.ValueString()
		createReq.ObservationID = &v
	}
	if !plan.ConfigID.IsNull() && !plan.ConfigID.IsUnknown() {
		v := plan.ConfigID.ValueString()
		createReq.ConfigID = &v
	}
	if !plan.Comment.IsNull() && !plan.Comment.IsUnknown() {
		v := plan.Comment.ValueString()
		createReq.Comment = &v
	}
	if !plan.Environment.IsNull() && !plan.Environment.IsUnknown() {
		v := plan.Environment.ValueString()
		createReq.Environment = &v
	}

	scoreID, err := r.client.CreateScore(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating score",
			fmt.Sprintf("Unable to create score, got error: %s", err),
		)
		return
	}

	s, err := r.client.GetScore(ctx, scoreID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading score after creation",
			fmt.Sprintf("Unable to read score %s after creation, got error: %s", scoreID, err),
		)
		return
	}

	applyScoreToModel(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scoreResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetScore(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading score",
			fmt.Sprintf("Unable to read score %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	applyScoreToModel(s, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *scoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Updating score",
		"All attributes of langfuse_score require replacement; in-place updates are not supported.",
	)
}

func (r *scoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state scoreResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteScore(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting score",
			fmt.Sprintf("Unable to delete score %s, got error: %s", state.ID.ValueString(), err),
		)
	}
}

func (r *scoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	s, err := r.client.GetScore(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing score",
			fmt.Sprintf("Unable to read score %s, got error: %s", req.ID, err),
		)
		return
	}

	var state scoreResourceModel
	applyScoreToModel(s, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
