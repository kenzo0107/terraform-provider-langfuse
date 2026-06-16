package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kenzo0107/terraform-provider-langfuse/langfuse"
)

var _ resource.Resource = &customModelResource{}
var _ resource.ResourceWithImportState = &customModelResource{}

func newCustomModelResource() resource.Resource {
	return &customModelResource{}
}

type customModelResource struct {
	client *langfuse.Client
}

type customModelResourceModel struct {
	ID                types.String  `tfsdk:"id"`
	ModelName         types.String  `tfsdk:"model_name"`
	MatchPattern      types.String  `tfsdk:"match_pattern"`
	InputPrice        types.Float64 `tfsdk:"input_price"`
	OutputPrice       types.Float64 `tfsdk:"output_price"`
	TotalPrice        types.Float64 `tfsdk:"total_price"`
	Unit              types.String  `tfsdk:"unit"`
	IsLangfuseManaged types.Bool    `tfsdk:"is_langfuse_managed"`
}

func (r *customModelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_model"
}

func (r *customModelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Langfuse custom model definition for cost tracking.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the custom model.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"model_name": schema.StringAttribute{
				MarkdownDescription: "The name of the custom model.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"match_pattern": schema.StringAttribute{
				MarkdownDescription: "The regex pattern to match model names.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"input_price": schema.Float64Attribute{
				MarkdownDescription: "Price per input token.",
				Optional:            true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
			},
			"output_price": schema.Float64Attribute{
				MarkdownDescription: "Price per output token.",
				Optional:            true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
			},
			"total_price": schema.Float64Attribute{
				MarkdownDescription: "Total price per token (alternative to input/output split).",
				Optional:            true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
			},
			"unit": schema.StringAttribute{
				MarkdownDescription: "The pricing unit, e.g. `TOKENS`.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_langfuse_managed": schema.BoolAttribute{
				MarkdownDescription: "Whether this model is managed by Langfuse (read-only).",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *customModelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func applyCustomModelToModel(m *langfuse.CustomModel, state *customModelResourceModel) {
	state.ID = types.StringValue(m.ID)
	state.ModelName = types.StringValue(m.ModelName)
	state.MatchPattern = types.StringValue(m.MatchPattern)
	state.IsLangfuseManaged = types.BoolValue(m.IsLangfuseManaged)

	if m.InputPrice != nil {
		state.InputPrice = types.Float64Value(*m.InputPrice)
	} else {
		state.InputPrice = types.Float64Null()
	}

	if m.OutputPrice != nil {
		state.OutputPrice = types.Float64Value(*m.OutputPrice)
	} else {
		state.OutputPrice = types.Float64Null()
	}

	if m.TotalPrice != nil {
		state.TotalPrice = types.Float64Value(*m.TotalPrice)
	} else {
		state.TotalPrice = types.Float64Null()
	}

	if m.Unit != nil {
		state.Unit = types.StringValue(*m.Unit)
	} else {
		state.Unit = types.StringNull()
	}
}

func (r *customModelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customModelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &langfuse.CreateCustomModelRequest{
		ModelName:    plan.ModelName.ValueString(),
		MatchPattern: plan.MatchPattern.ValueString(),
	}
	if !plan.InputPrice.IsNull() && !plan.InputPrice.IsUnknown() {
		v := plan.InputPrice.ValueFloat64()
		createReq.InputPrice = &v
	}
	if !plan.OutputPrice.IsNull() && !plan.OutputPrice.IsUnknown() {
		v := plan.OutputPrice.ValueFloat64()
		createReq.OutputPrice = &v
	}
	if !plan.TotalPrice.IsNull() && !plan.TotalPrice.IsUnknown() {
		v := plan.TotalPrice.ValueFloat64()
		createReq.TotalPrice = &v
	}
	if !plan.Unit.IsNull() && !plan.Unit.IsUnknown() {
		v := plan.Unit.ValueString()
		createReq.Unit = &v
	}

	m, err := r.client.CreateCustomModel(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating custom model",
			fmt.Sprintf("Unable to create custom model, got error: %s", err),
		)
		return
	}

	applyCustomModelToModel(m, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customModelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customModelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.GetCustomModel(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading custom model",
			fmt.Sprintf("Unable to read custom model %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	applyCustomModelToModel(m, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *customModelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes have RequiresReplace, so Update should never be called.
	resp.Diagnostics.AddError(
		"Updating custom model",
		"Custom model does not support in-place updates. All changes require replacement.",
	)
}

func (r *customModelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customModelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCustomModel(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting custom model",
			fmt.Sprintf("Unable to delete custom model %s, got error: %s", state.ID.ValueString(), err),
		)
	}
}

func (r *customModelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	m, err := r.client.GetCustomModel(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing custom model",
			fmt.Sprintf("Unable to read custom model %s, got error: %s", req.ID, err),
		)
		return
	}

	var state customModelResourceModel
	applyCustomModelToModel(m, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
