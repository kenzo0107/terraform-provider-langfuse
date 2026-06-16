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

var _ resource.Resource = &evaluatorResource{}
var _ resource.ResourceWithImportState = &evaluatorResource{}

func newEvaluatorResource() resource.Resource {
	return &evaluatorResource{}
}

type evaluatorResource struct {
	client *langfuse.Client
}

type evaluatorResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Prompt     types.String `tfsdk:"prompt"`
	SourceCode types.String `tfsdk:"source_code"`
}

func (r *evaluatorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_evaluator"
}

func (r *evaluatorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Langfuse evaluator.\n\n> **Warning:** This resource uses the unstable Langfuse API (`/api/public/unstable/`) and may change without notice.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the evaluator.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the evaluator.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the evaluator. Must be `llm_as_judge` or `code`. Changing this creates a new evaluator.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prompt": schema.StringAttribute{
				MarkdownDescription: "A JSON string representing the prompt configuration (for `type = \"llm_as_judge\"`).",
				Optional:            true,
			},
			"source_code": schema.StringAttribute{
				MarkdownDescription: "The source code of the evaluator (for `type = \"code\"`).",
				Optional:            true,
			},
		},
	}
}

func (r *evaluatorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func applyEvaluatorToModel(ev *langfuse.Evaluator, model *evaluatorResourceModel) {
	model.ID = types.StringValue(ev.ID)
	model.Name = types.StringValue(ev.Name)
	model.Type = types.StringValue(ev.Type)

	if ev.Prompt != nil {
		model.Prompt = types.StringValue(*ev.Prompt)
	} else {
		model.Prompt = types.StringNull()
	}
	if ev.SourceCode != nil {
		model.SourceCode = types.StringValue(*ev.SourceCode)
	} else {
		model.SourceCode = types.StringNull()
	}
}

func (r *evaluatorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan evaluatorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &langfuse.CreateEvaluatorRequest{
		Name: plan.Name.ValueString(),
		Type: plan.Type.ValueString(),
	}
	if !plan.Prompt.IsNull() && !plan.Prompt.IsUnknown() {
		v := plan.Prompt.ValueString()
		createReq.Prompt = &v
	}
	if !plan.SourceCode.IsNull() && !plan.SourceCode.IsUnknown() {
		v := plan.SourceCode.ValueString()
		createReq.SourceCode = &v
	}

	ev, err := r.client.CreateEvaluator(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating evaluator",
			fmt.Sprintf("Unable to create evaluator, got error: %s", err),
		)
		return
	}

	applyEvaluatorToModel(ev, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *evaluatorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state evaluatorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ev, err := r.client.GetEvaluator(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading evaluator",
			fmt.Sprintf("Unable to read evaluator %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	applyEvaluatorToModel(ev, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *evaluatorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state evaluatorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &langfuse.UpdateEvaluatorRequest{}
	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		updateReq.Name = &v
	}
	if !plan.Prompt.IsNull() && !plan.Prompt.IsUnknown() {
		v := plan.Prompt.ValueString()
		updateReq.Prompt = &v
	}
	if !plan.SourceCode.IsNull() && !plan.SourceCode.IsUnknown() {
		v := plan.SourceCode.ValueString()
		updateReq.SourceCode = &v
	}

	ev, err := r.client.UpdateEvaluator(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating evaluator",
			fmt.Sprintf("Unable to update evaluator %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	applyEvaluatorToModel(ev, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *evaluatorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state evaluatorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteEvaluator(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting evaluator",
			fmt.Sprintf("Unable to delete evaluator %s, got error: %s", state.ID.ValueString(), err),
		)
	}
}

func (r *evaluatorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ev, err := r.client.GetEvaluator(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing evaluator",
			fmt.Sprintf("Unable to read evaluator %s, got error: %s", req.ID, err),
		)
		return
	}

	var state evaluatorResourceModel
	applyEvaluatorToModel(ev, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
