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

var _ resource.Resource = &datasetItemResource{}
var _ resource.ResourceWithImportState = &datasetItemResource{}

func newDatasetItemResource() resource.Resource {
	return &datasetItemResource{}
}

type datasetItemResource struct {
	client *langfuse.Client
}

type datasetItemResourceModel struct {
	ID             types.String `tfsdk:"id"`
	DatasetName    types.String `tfsdk:"dataset_name"`
	Input          types.String `tfsdk:"input"`
	ExpectedOutput types.String `tfsdk:"expected_output"`
	Status         types.String `tfsdk:"status"`
}

func (r *datasetItemResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset_item"
}

func (r *datasetItemResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Langfuse dataset item.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the dataset item.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dataset_name": schema.StringAttribute{
				MarkdownDescription: "The name of the dataset this item belongs to. Changing this creates a new item.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"input": schema.StringAttribute{
				MarkdownDescription: "The input for the dataset item as a JSON string. Changing this creates a new item.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expected_output": schema.StringAttribute{
				MarkdownDescription: "The expected output for the dataset item as a JSON string. Changing this creates a new item.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the dataset item. Must be `ACTIVE` or `ARCHIVED`. Changing this creates a new item.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *datasetItemResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *datasetItemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan datasetItemResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &langfuse.CreateDatasetItemRequest{
		DatasetName: plan.DatasetName.ValueString(),
	}
	if !plan.Input.IsNull() && !plan.Input.IsUnknown() {
		v := plan.Input.ValueString()
		createReq.Input = &v
	}
	if !plan.ExpectedOutput.IsNull() && !plan.ExpectedOutput.IsUnknown() {
		v := plan.ExpectedOutput.ValueString()
		createReq.ExpectedOutput = &v
	}
	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		createReq.Status = plan.Status.ValueString()
	}

	item, err := r.client.CreateDatasetItem(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating dataset item",
			fmt.Sprintf("Unable to create dataset item, got error: %s", err),
		)
		return
	}

	plan.ID = types.StringValue(item.ID)
	plan.DatasetName = types.StringValue(item.DatasetName)
	plan.Status = types.StringValue(item.Status)
	if item.Input != nil {
		plan.Input = types.StringValue(*item.Input)
	}
	if item.ExpectedOutput != nil {
		plan.ExpectedOutput = types.StringValue(*item.ExpectedOutput)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *datasetItemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state datasetItemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := r.client.GetDatasetItem(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading dataset item",
			fmt.Sprintf("Unable to read dataset item %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.DatasetName = types.StringValue(item.DatasetName)
	state.Status = types.StringValue(item.Status)
	if item.Input != nil {
		state.Input = types.StringValue(*item.Input)
	} else {
		state.Input = types.StringNull()
	}
	if item.ExpectedOutput != nil {
		state.ExpectedOutput = types.StringValue(*item.ExpectedOutput)
	} else {
		state.ExpectedOutput = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datasetItemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Updating dataset item",
		"All attributes of langfuse_dataset_item require replacement; in-place updates are not supported.",
	)
}

func (r *datasetItemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state datasetItemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDatasetItem(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting dataset item",
			fmt.Sprintf("Unable to delete dataset item %s, got error: %s", state.ID.ValueString(), err),
		)
	}
}

func (r *datasetItemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	itemID := req.ID

	item, err := r.client.GetDatasetItem(ctx, itemID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing dataset item",
			fmt.Sprintf("Unable to read dataset item %s, got error: %s", itemID, err),
		)
		return
	}

	state := datasetItemResourceModel{
		ID:          types.StringValue(item.ID),
		DatasetName: types.StringValue(item.DatasetName),
		Status:      types.StringValue(item.Status),
	}
	if item.Input != nil {
		state.Input = types.StringValue(*item.Input)
	} else {
		state.Input = types.StringNull()
	}
	if item.ExpectedOutput != nil {
		state.ExpectedOutput = types.StringValue(*item.ExpectedOutput)
	} else {
		state.ExpectedOutput = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
