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

var _ resource.Resource = &datasetResource{}
var _ resource.ResourceWithImportState = &datasetResource{}

func newDatasetResource() resource.Resource {
	return &datasetResource{}
}

type datasetResource struct {
	client *langfuse.Client
}

type datasetResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (r *datasetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset"
}

func (r *datasetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Langfuse dataset.\n\n> **Note:** The Langfuse API does not support deleting datasets. Destroying this resource only removes it from Terraform state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the dataset.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The project ID the dataset belongs to.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the dataset. Changing this creates a new dataset.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "An optional description for the dataset.",
				Optional:            true,
			},
		},
	}
}

func (r *datasetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *datasetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan datasetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &langfuse.CreateDatasetRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		createReq.Description = &v
	}

	ds, err := r.client.CreateDataset(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating dataset",
			fmt.Sprintf("Unable to create dataset, got error: %s", err),
		)
		return
	}

	plan.ID = types.StringValue(ds.ID)
	plan.ProjectID = types.StringValue(ds.ProjectID)
	plan.Name = types.StringValue(ds.Name)
	if ds.Description != nil {
		plan.Description = types.StringValue(*ds.Description)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *datasetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state datasetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ds, err := r.client.GetDataset(ctx, state.Name.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading dataset",
			fmt.Sprintf("Unable to read dataset %s, got error: %s", state.Name.ValueString(), err),
		)
		return
	}

	state.ID = types.StringValue(ds.ID)
	state.ProjectID = types.StringValue(ds.ProjectID)
	state.Name = types.StringValue(ds.Name)
	if ds.Description != nil {
		state.Description = types.StringValue(*ds.Description)
	} else {
		state.Description = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datasetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan datasetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &langfuse.CreateDatasetRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		createReq.Description = &v
	}

	ds, err := r.client.CreateDataset(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating dataset",
			fmt.Sprintf("Unable to update dataset %s, got error: %s", plan.Name.ValueString(), err),
		)
		return
	}

	plan.ID = types.StringValue(ds.ID)
	plan.ProjectID = types.StringValue(ds.ProjectID)
	plan.Name = types.StringValue(ds.Name)
	if ds.Description != nil {
		plan.Description = types.StringValue(*ds.Description)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *datasetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Langfuse has no delete API for datasets; removing from state only.
}

func (r *datasetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	ds, err := r.client.GetDataset(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing dataset",
			fmt.Sprintf("Unable to read dataset %s, got error: %s", name, err),
		)
		return
	}

	state := datasetResourceModel{
		ID:        types.StringValue(ds.ID),
		ProjectID: types.StringValue(ds.ProjectID),
		Name:      types.StringValue(ds.Name),
	}
	if ds.Description != nil {
		state.Description = types.StringValue(*ds.Description)
	} else {
		state.Description = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
