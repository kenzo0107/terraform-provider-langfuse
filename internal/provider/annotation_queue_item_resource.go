package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kenzo0107/terraform-provider-langfuse/langfuse"
)

var _ resource.Resource = &annotationQueueItemResource{}
var _ resource.ResourceWithImportState = &annotationQueueItemResource{}

func newAnnotationQueueItemResource() resource.Resource {
	return &annotationQueueItemResource{}
}

type annotationQueueItemResource struct {
	client *langfuse.Client
}

type annotationQueueItemResourceModel struct {
	ID            types.String `tfsdk:"id"`
	QueueID       types.String `tfsdk:"queue_id"`
	TraceID       types.String `tfsdk:"trace_id"`
	ObservationID types.String `tfsdk:"observation_id"`
	Status        types.String `tfsdk:"status"`
}

func (r *annotationQueueItemResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_annotation_queue_item"
}

func (r *annotationQueueItemResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Langfuse annotation queue item (a trace or observation added to an annotation queue).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the queue item.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"queue_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the annotation queue. Changing this creates a new item.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"trace_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the trace to annotate. Changing this creates a new item.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"observation_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the observation to annotate (optional). Changing this creates a new item.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the queue item. Must be `QUEUED`, `ACTIVE`, or `COMPLETED`.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *annotationQueueItemResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func applyAnnotationQueueItemToModel(item *langfuse.AnnotationQueueItem, model *annotationQueueItemResourceModel) {
	model.ID = types.StringValue(item.ID)
	model.QueueID = types.StringValue(item.QueueID)
	model.TraceID = types.StringValue(item.TraceID)
	model.Status = types.StringValue(item.Status)
	if item.ObservationID != nil {
		model.ObservationID = types.StringValue(*item.ObservationID)
	} else {
		model.ObservationID = types.StringNull()
	}
}

func (r *annotationQueueItemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan annotationQueueItemResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &langfuse.CreateAnnotationQueueItemRequest{
		TraceID: plan.TraceID.ValueString(),
	}
	if !plan.ObservationID.IsNull() && !plan.ObservationID.IsUnknown() {
		v := plan.ObservationID.ValueString()
		createReq.ObservationID = &v
	}

	item, err := r.client.CreateAnnotationQueueItem(ctx, plan.QueueID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating annotation queue item",
			fmt.Sprintf("Unable to create annotation queue item, got error: %s", err),
		)
		return
	}

	applyAnnotationQueueItemToModel(item, &plan)

	if !plan.Status.IsNull() && !plan.Status.IsUnknown() && plan.Status.ValueString() != item.Status {
		updated, err := r.client.UpdateAnnotationQueueItem(ctx, plan.QueueID.ValueString(), item.ID, &langfuse.UpdateAnnotationQueueItemRequest{
			Status: plan.Status.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Updating annotation queue item status",
				fmt.Sprintf("Unable to set status for annotation queue item, got error: %s", err),
			)
			return
		}
		applyAnnotationQueueItemToModel(updated, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *annotationQueueItemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state annotationQueueItemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := r.client.GetAnnotationQueueItem(ctx, state.QueueID.ValueString(), state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading annotation queue item",
			fmt.Sprintf("Unable to read annotation queue item %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	applyAnnotationQueueItemToModel(item, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *annotationQueueItemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state annotationQueueItemResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := r.client.UpdateAnnotationQueueItem(ctx, state.QueueID.ValueString(), state.ID.ValueString(), &langfuse.UpdateAnnotationQueueItemRequest{
		Status: plan.Status.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating annotation queue item",
			fmt.Sprintf("Unable to update annotation queue item %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	applyAnnotationQueueItemToModel(item, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *annotationQueueItemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state annotationQueueItemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAnnotationQueueItem(ctx, state.QueueID.ValueString(), state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting annotation queue item",
			fmt.Sprintf("Unable to delete annotation queue item %s, got error: %s", state.ID.ValueString(), err),
		)
	}
}

func (r *annotationQueueItemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format `{queue_id}/{item_id}`.",
		)
		return
	}

	queueID, itemID := parts[0], parts[1]

	item, err := r.client.GetAnnotationQueueItem(ctx, queueID, itemID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing annotation queue item",
			fmt.Sprintf("Unable to read annotation queue item %s, got error: %s", itemID, err),
		)
		return
	}

	var state annotationQueueItemResourceModel
	applyAnnotationQueueItemToModel(item, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
