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

var _ resource.Resource = &evaluationRuleResource{}
var _ resource.ResourceWithImportState = &evaluationRuleResource{}

func newEvaluationRuleResource() resource.Resource {
	return &evaluationRuleResource{}
}

type evaluationRuleResource struct {
	client *langfuse.Client
}

type evaluationRuleResourceModel struct {
	ID          types.String  `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	State       types.String  `tfsdk:"state"`
	Target      types.String  `tfsdk:"target"`
	EvaluatorID types.String  `tfsdk:"evaluator_id"`
	Filter      types.String  `tfsdk:"filter"`
	Mapping     types.String  `tfsdk:"mapping"`
	Sampling    types.Float64 `tfsdk:"sampling"`
	Priority    types.Int64   `tfsdk:"priority"`
}

func (r *evaluationRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_evaluation_rule"
}

func (r *evaluationRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Langfuse evaluation rule.\n\n> **Warning:** This resource uses the unstable Langfuse API (`/api/public/unstable/`) and may change without notice.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the evaluation rule.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the evaluation rule.",
				Required:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The state of the rule. Must be `ACTIVE` or `INACTIVE`.",
				Optional:            true,
				Computed:            true,
			},
			"target": schema.StringAttribute{
				MarkdownDescription: "The target type for the rule. Must be `TRACE` or `DATASET_RUN`. Changing this creates a new rule.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"evaluator_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the evaluator to use. Changing this creates a new rule.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"filter": schema.StringAttribute{
				MarkdownDescription: "A JSON string representing filter conditions for the rule.",
				Optional:            true,
			},
			"mapping": schema.StringAttribute{
				MarkdownDescription: "A JSON string representing variable mappings for the evaluator.",
				Optional:            true,
			},
			"sampling": schema.Float64Attribute{
				MarkdownDescription: "The sampling rate for the rule (0.0 to 1.0).",
				Optional:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "The execution priority of the rule.",
				Optional:            true,
			},
		},
	}
}

func (r *evaluationRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func applyEvaluationRuleToModel(rule *langfuse.EvaluationRule, model *evaluationRuleResourceModel) {
	model.ID = types.StringValue(rule.ID)
	model.Name = types.StringValue(rule.Name)
	model.State = types.StringValue(rule.State)
	model.Target = types.StringValue(rule.Target)
	model.EvaluatorID = types.StringValue(rule.EvaluatorID)

	if rule.Filter != nil {
		model.Filter = types.StringValue(*rule.Filter)
	} else {
		model.Filter = types.StringNull()
	}
	if rule.Mapping != nil {
		model.Mapping = types.StringValue(*rule.Mapping)
	} else {
		model.Mapping = types.StringNull()
	}
	if rule.Sampling != nil {
		model.Sampling = types.Float64Value(*rule.Sampling)
	} else {
		model.Sampling = types.Float64Null()
	}
	if rule.Priority != nil {
		model.Priority = types.Int64Value(int64(*rule.Priority))
	} else {
		model.Priority = types.Int64Null()
	}
}

func (r *evaluationRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan evaluationRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &langfuse.CreateEvaluationRuleRequest{
		Name:        plan.Name.ValueString(),
		Target:      plan.Target.ValueString(),
		EvaluatorID: plan.EvaluatorID.ValueString(),
	}
	if !plan.Filter.IsNull() && !plan.Filter.IsUnknown() {
		v := plan.Filter.ValueString()
		createReq.Filter = &v
	}
	if !plan.Mapping.IsNull() && !plan.Mapping.IsUnknown() {
		v := plan.Mapping.ValueString()
		createReq.Mapping = &v
	}
	if !plan.Sampling.IsNull() && !plan.Sampling.IsUnknown() {
		v := plan.Sampling.ValueFloat64()
		createReq.Sampling = &v
	}
	if !plan.Priority.IsNull() && !plan.Priority.IsUnknown() {
		v := int(plan.Priority.ValueInt64())
		createReq.Priority = &v
	}

	rule, err := r.client.CreateEvaluationRule(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating evaluation rule",
			fmt.Sprintf("Unable to create evaluation rule, got error: %s", err),
		)
		return
	}

	applyEvaluationRuleToModel(rule, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *evaluationRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state evaluationRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.GetEvaluationRule(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading evaluation rule",
			fmt.Sprintf("Unable to read evaluation rule %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	applyEvaluationRuleToModel(rule, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *evaluationRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state evaluationRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &langfuse.UpdateEvaluationRuleRequest{}
	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		updateReq.Name = &v
	}
	if !plan.State.Equal(state.State) {
		v := plan.State.ValueString()
		updateReq.State = &v
	}
	if !plan.Filter.IsNull() && !plan.Filter.IsUnknown() {
		v := plan.Filter.ValueString()
		updateReq.Filter = &v
	}
	if !plan.Mapping.IsNull() && !plan.Mapping.IsUnknown() {
		v := plan.Mapping.ValueString()
		updateReq.Mapping = &v
	}
	if !plan.Sampling.IsNull() && !plan.Sampling.IsUnknown() {
		v := plan.Sampling.ValueFloat64()
		updateReq.Sampling = &v
	}
	if !plan.Priority.IsNull() && !plan.Priority.IsUnknown() {
		v := int(plan.Priority.ValueInt64())
		updateReq.Priority = &v
	}

	rule, err := r.client.UpdateEvaluationRule(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating evaluation rule",
			fmt.Sprintf("Unable to update evaluation rule %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	applyEvaluationRuleToModel(rule, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *evaluationRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state evaluationRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteEvaluationRule(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting evaluation rule",
			fmt.Sprintf("Unable to delete evaluation rule %s, got error: %s", state.ID.ValueString(), err),
		)
	}
}

func (r *evaluationRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	rule, err := r.client.GetEvaluationRule(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing evaluation rule",
			fmt.Sprintf("Unable to read evaluation rule %s, got error: %s", req.ID, err),
		)
		return
	}

	var state evaluationRuleResourceModel
	applyEvaluationRuleToModel(rule, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
