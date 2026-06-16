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

var _ resource.Resource = &scoreConfigResource{}
var _ resource.ResourceWithImportState = &scoreConfigResource{}

func newScoreConfigResource() resource.Resource {
	return &scoreConfigResource{}
}

type scoreConfigResource struct {
	client *langfuse.Client
}

type scoreConfigResourceModel struct {
	ID          types.String  `tfsdk:"id"`
	ProjectID   types.String  `tfsdk:"project_id"`
	Name        types.String  `tfsdk:"name"`
	DataType    types.String  `tfsdk:"data_type"`
	IsArchived  types.Bool    `tfsdk:"is_archived"`
	MinValue    types.Float64 `tfsdk:"min_value"`
	MaxValue    types.Float64 `tfsdk:"max_value"`
	Description types.String  `tfsdk:"description"`
	Categories  types.List    `tfsdk:"categories"`
}

type categoryModel struct {
	Value types.Float64 `tfsdk:"value"`
	Label types.String  `tfsdk:"label"`
}

var categoryAttrTypes = map[string]attr.Type{
	"value": types.Float64Type,
	"label": types.StringType,
}

func (r *scoreConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_score_config"
}

func (r *scoreConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Langfuse score configuration. Score configs define the evaluation criteria for LLM outputs. " +
			"Note: The Langfuse API does not support deleting score configs; destroying this resource will archive it instead.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the score config.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project this score config belongs to (read-only, set by Langfuse).",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the score config.",
				Required:            true,
			},
			"data_type": schema.StringAttribute{
				MarkdownDescription: "The data type for scores. Must be one of `NUMERIC`, `BOOLEAN`, or `CATEGORICAL`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_archived": schema.BoolAttribute{
				MarkdownDescription: "Whether the score config is archived.",
				Computed:            true,
			},
			"min_value": schema.Float64Attribute{
				MarkdownDescription: "The minimum allowed value (for `NUMERIC` data type).",
				Optional:            true,
			},
			"max_value": schema.Float64Attribute{
				MarkdownDescription: "The maximum allowed value (for `NUMERIC` data type).",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the score config.",
				Optional:            true,
			},
			"categories": schema.ListNestedAttribute{
				MarkdownDescription: "Category definitions (for `CATEGORICAL` data type).",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"value": schema.Float64Attribute{
							MarkdownDescription: "The numeric value for this category.",
							Required:            true,
						},
						"label": schema.StringAttribute{
							MarkdownDescription: "The label for this category.",
							Required:            true,
						},
					},
				},
			},
		},
	}
}

func (r *scoreConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func categoriesFromModel(ctx context.Context, list types.List) ([]*langfuse.ConfigCategory, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	var models []categoryModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, fmt.Errorf("reading categories from plan")
	}

	cats := make([]*langfuse.ConfigCategory, len(models))
	for i, c := range models {
		cats[i] = &langfuse.ConfigCategory{
			Value: c.Value.ValueFloat64(),
			Label: c.Label.ValueString(),
		}
	}
	return cats, nil
}

func categoriesToList(ctx context.Context, cats []*langfuse.ConfigCategory) (types.List, error) {
	if len(cats) == 0 {
		return types.ListValueMust(types.ObjectType{AttrTypes: categoryAttrTypes}, []attr.Value{}), nil
	}

	elems := make([]attr.Value, len(cats))
	for i, c := range cats {
		obj, diags := types.ObjectValue(categoryAttrTypes, map[string]attr.Value{
			"value": types.Float64Value(c.Value),
			"label": types.StringValue(c.Label),
		})
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: categoryAttrTypes}), fmt.Errorf("building category object")
		}
		elems[i] = obj
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: categoryAttrTypes}, elems)
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: categoryAttrTypes}), fmt.Errorf("building categories list")
	}
	return list, nil
}

func applyScoreConfigToModel(ctx context.Context, sc *langfuse.ScoreConfig, state *scoreConfigResourceModel) error {
	state.ID = types.StringValue(sc.ID)
	state.ProjectID = types.StringValue(sc.ProjectID)
	state.Name = types.StringValue(sc.Name)
	state.DataType = types.StringValue(string(sc.DataType))
	state.IsArchived = types.BoolValue(sc.IsArchived)

	if sc.MinValue != nil {
		state.MinValue = types.Float64Value(*sc.MinValue)
	} else {
		state.MinValue = types.Float64Null()
	}

	if sc.MaxValue != nil {
		state.MaxValue = types.Float64Value(*sc.MaxValue)
	} else {
		state.MaxValue = types.Float64Null()
	}

	if sc.Description != nil {
		state.Description = types.StringValue(*sc.Description)
	} else {
		state.Description = types.StringNull()
	}

	cats, err := categoriesToList(ctx, sc.Categories)
	if err != nil {
		return err
	}
	state.Categories = cats
	return nil
}

func (r *scoreConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan scoreConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cats, err := categoriesFromModel(ctx, plan.Categories)
	if err != nil {
		resp.Diagnostics.AddError("Creating score config", err.Error())
		return
	}

	createReq := &langfuse.CreateScoreConfigRequest{
		Name:       plan.Name.ValueString(),
		DataType:   langfuse.ScoreConfigDataType(plan.DataType.ValueString()),
		Categories: cats,
	}
	if !plan.MinValue.IsNull() && !plan.MinValue.IsUnknown() {
		v := plan.MinValue.ValueFloat64()
		createReq.MinValue = &v
	}
	if !plan.MaxValue.IsNull() && !plan.MaxValue.IsUnknown() {
		v := plan.MaxValue.ValueFloat64()
		createReq.MaxValue = &v
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		createReq.Description = &v
	}

	sc, err := r.client.CreateScoreConfig(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating score config",
			fmt.Sprintf("Unable to create score config, got error: %s", err),
		)
		return
	}

	if err := applyScoreConfigToModel(ctx, sc, &plan); err != nil {
		resp.Diagnostics.AddError("Creating score config", fmt.Sprintf("Unable to map response: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scoreConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scoreConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sc, err := r.client.GetScoreConfig(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading score config",
			fmt.Sprintf("Unable to read score config %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	if err := applyScoreConfigToModel(ctx, sc, &state); err != nil {
		resp.Diagnostics.AddError("Reading score config", fmt.Sprintf("Unable to map response: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *scoreConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state scoreConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cats, err := categoriesFromModel(ctx, plan.Categories)
	if err != nil {
		resp.Diagnostics.AddError("Updating score config", err.Error())
		return
	}

	name := plan.Name.ValueString()
	updateReq := &langfuse.UpdateScoreConfigRequest{
		Name:       &name,
		Categories: cats,
	}
	if !plan.MinValue.IsNull() && !plan.MinValue.IsUnknown() {
		v := plan.MinValue.ValueFloat64()
		updateReq.MinValue = &v
	}
	if !plan.MaxValue.IsNull() && !plan.MaxValue.IsUnknown() {
		v := plan.MaxValue.ValueFloat64()
		updateReq.MaxValue = &v
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		updateReq.Description = &v
	}

	sc, err := r.client.UpdateScoreConfig(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating score config",
			fmt.Sprintf("Unable to update score config %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	if err := applyScoreConfigToModel(ctx, sc, &state); err != nil {
		resp.Diagnostics.AddError("Updating score config", fmt.Sprintf("Unable to map response: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *scoreConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state scoreConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.ArchiveScoreConfig(ctx, state.ID.ValueString()); err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting score config",
			fmt.Sprintf("Unable to archive score config %s, got error: %s", state.ID.ValueString(), err),
		)
	}
}

func (r *scoreConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	sc, err := r.client.GetScoreConfig(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing score config",
			fmt.Sprintf("Unable to read score config %s, got error: %s", req.ID, err),
		)
		return
	}

	var state scoreConfigResourceModel
	if err := applyScoreConfigToModel(ctx, sc, &state); err != nil {
		resp.Diagnostics.AddError("Importing score config", fmt.Sprintf("Unable to map response: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
