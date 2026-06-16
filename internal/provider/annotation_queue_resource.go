package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kenzo0107/terraform-provider-langfuse/langfuse"
)

var _ resource.Resource = &annotationQueueResource{}
var _ resource.ResourceWithImportState = &annotationQueueResource{}

func newAnnotationQueueResource() resource.Resource {
	return &annotationQueueResource{}
}

type annotationQueueResource struct {
	client *langfuse.Client
}

type annotationQueueResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProjectID      types.String `tfsdk:"project_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	ScoreConfigIDs types.List   `tfsdk:"score_config_ids"`
}

func (r *annotationQueueResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_annotation_queue"
}

func (r *annotationQueueResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Langfuse annotation queue.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the annotation queue.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project this annotation queue belongs to (read-only, set by Langfuse).",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the annotation queue.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the annotation queue.",
				Optional:            true,
			},
			"score_config_ids": schema.ListAttribute{
				MarkdownDescription: "List of score config IDs associated with this annotation queue.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *annotationQueueResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func scoreConfigIDsFromList(ctx context.Context, list types.List) ([]string, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	var ids []string
	if diags := list.ElementsAs(ctx, &ids, false); diags.HasError() {
		return nil, fmt.Errorf("reading score_config_ids from plan")
	}
	return ids, nil
}

func scoreConfigIDsToList(ids []string) types.List {
	if len(ids) == 0 {
		return types.ListValueMust(types.StringType, []attr.Value{})
	}

	elems := make([]attr.Value, len(ids))
	for i, id := range ids {
		elems[i] = types.StringValue(id)
	}
	return types.ListValueMust(types.StringType, elems)
}

func applyAnnotationQueueToModel(q *langfuse.AnnotationQueue, state *annotationQueueResourceModel) {
	state.ID = types.StringValue(q.ID)
	state.ProjectID = types.StringValue(q.ProjectID)
	state.Name = types.StringValue(q.Name)

	if q.Description != nil {
		state.Description = types.StringValue(*q.Description)
	} else {
		state.Description = types.StringNull()
	}

	state.ScoreConfigIDs = scoreConfigIDsToList(q.ScoreConfigIDs)
}

func (r *annotationQueueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan annotationQueueResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scoreConfigIDs, err := scoreConfigIDsFromList(ctx, plan.ScoreConfigIDs)
	if err != nil {
		resp.Diagnostics.AddError("Creating annotation queue", err.Error())
		return
	}

	createReq := &langfuse.CreateAnnotationQueueRequest{
		Name:           plan.Name.ValueString(),
		ScoreConfigIDs: scoreConfigIDs,
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		createReq.Description = &v
	}

	q, err := r.client.CreateAnnotationQueue(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating annotation queue",
			fmt.Sprintf("Unable to create annotation queue, got error: %s", err),
		)
		return
	}

	applyAnnotationQueueToModel(q, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *annotationQueueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state annotationQueueResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	q, err := r.client.GetAnnotationQueue(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading annotation queue",
			fmt.Sprintf("Unable to read annotation queue %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	applyAnnotationQueueToModel(q, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *annotationQueueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state annotationQueueResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scoreConfigIDs, err := scoreConfigIDsFromList(ctx, plan.ScoreConfigIDs)
	if err != nil {
		resp.Diagnostics.AddError("Updating annotation queue", err.Error())
		return
	}

	name := plan.Name.ValueString()
	updateReq := &langfuse.UpdateAnnotationQueueRequest{
		Name:           &name,
		ScoreConfigIDs: scoreConfigIDs,
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		updateReq.Description = &v
	}

	q, err := r.client.UpdateAnnotationQueue(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating annotation queue",
			fmt.Sprintf("Unable to update annotation queue %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	applyAnnotationQueueToModel(q, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *annotationQueueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state annotationQueueResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAnnotationQueue(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting annotation queue",
			fmt.Sprintf("Unable to delete annotation queue %s, got error: %s", state.ID.ValueString(), err),
		)
	}
}

func (r *annotationQueueResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	q, err := r.client.GetAnnotationQueue(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing annotation queue",
			fmt.Sprintf("Unable to read annotation queue %s, got error: %s", req.ID, err),
		)
		return
	}

	var state annotationQueueResourceModel
	applyAnnotationQueueToModel(q, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
