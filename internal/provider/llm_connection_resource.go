package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kenzo0107/terraform-provider-langfuse/langfuse"
)

var _ resource.Resource = &llmConnectionResource{}
var _ resource.ResourceWithImportState = &llmConnectionResource{}

func newLLMConnectionResource() resource.Resource {
	return &llmConnectionResource{}
}

type llmConnectionResource struct {
	client *langfuse.Client
}

type llmConnectionResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Provider          types.String `tfsdk:"provider"`
	BaseURL           types.String `tfsdk:"base_url"`
	APIKey            types.String `tfsdk:"api_key"`
	WithDefaultModels types.Bool   `tfsdk:"with_default_models"`
}

func (r *llmConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_llm_connection"
}

func (r *llmConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Langfuse LLM connection for playground and evaluation features.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the LLM connection.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the LLM connection.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provider": schema.StringAttribute{
				MarkdownDescription: "The LLM provider (e.g. `openai`, `anthropic`).",
				Required:            true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "The base URL for the LLM provider API.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key for the LLM provider. This is write-only and will not be read back from the API.",
				Optional:            true,
				Sensitive:           true,
			},
			"with_default_models": schema.BoolAttribute{
				MarkdownDescription: "Whether to include default models for this provider.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *llmConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *llmConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan llmConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	upsertReq := &langfuse.UpsertLLMConnectionRequest{
		Name:     plan.Name.ValueString(),
		Provider: plan.Provider.ValueString(),
	}
	if !plan.BaseURL.IsNull() && !plan.BaseURL.IsUnknown() {
		v := plan.BaseURL.ValueString()
		upsertReq.BaseURL = &v
	}
	if !plan.APIKey.IsNull() && !plan.APIKey.IsUnknown() {
		v := plan.APIKey.ValueString()
		upsertReq.APIKey = &v
	}
	if !plan.WithDefaultModels.IsNull() && !plan.WithDefaultModels.IsUnknown() {
		v := plan.WithDefaultModels.ValueBool()
		upsertReq.WithDefaultModels = &v
	}

	conn, err := r.client.UpsertLLMConnection(ctx, upsertReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating LLM connection",
			fmt.Sprintf("Unable to create LLM connection, got error: %s", err),
		)
		return
	}

	plan.ID = types.StringValue(conn.ID)
	plan.Name = types.StringValue(conn.Name)
	plan.Provider = types.StringValue(conn.Provider)
	plan.WithDefaultModels = types.BoolValue(conn.WithDefaultModels)

	if conn.BaseURL != nil {
		plan.BaseURL = types.StringValue(*conn.BaseURL)
	} else {
		plan.BaseURL = types.StringNull()
	}
	// api_key is intentionally not set from the response (write-only).

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *llmConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state llmConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn, err := r.client.GetLLMConnectionByName(ctx, state.Name.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading LLM connection",
			fmt.Sprintf("Unable to read LLM connection %s, got error: %s", state.Name.ValueString(), err),
		)
		return
	}

	state.ID = types.StringValue(conn.ID)
	state.Name = types.StringValue(conn.Name)
	state.Provider = types.StringValue(conn.Provider)
	state.WithDefaultModels = types.BoolValue(conn.WithDefaultModels)

	if conn.BaseURL != nil {
		state.BaseURL = types.StringValue(*conn.BaseURL)
	} else {
		state.BaseURL = types.StringNull()
	}
	// api_key is not read back from the API; keep whatever is in state.

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *llmConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state llmConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	upsertReq := &langfuse.UpsertLLMConnectionRequest{
		Name:     state.Name.ValueString(),
		Provider: plan.Provider.ValueString(),
	}
	if !plan.BaseURL.IsNull() && !plan.BaseURL.IsUnknown() {
		v := plan.BaseURL.ValueString()
		upsertReq.BaseURL = &v
	}
	if !plan.APIKey.IsNull() && !plan.APIKey.IsUnknown() {
		v := plan.APIKey.ValueString()
		upsertReq.APIKey = &v
	}
	if !plan.WithDefaultModels.IsNull() && !plan.WithDefaultModels.IsUnknown() {
		v := plan.WithDefaultModels.ValueBool()
		upsertReq.WithDefaultModels = &v
	}

	conn, err := r.client.UpsertLLMConnection(ctx, upsertReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating LLM connection",
			fmt.Sprintf("Unable to update LLM connection %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.ID = types.StringValue(conn.ID)
	state.Name = types.StringValue(conn.Name)
	state.Provider = types.StringValue(conn.Provider)
	state.WithDefaultModels = types.BoolValue(conn.WithDefaultModels)

	if conn.BaseURL != nil {
		state.BaseURL = types.StringValue(*conn.BaseURL)
	} else {
		state.BaseURL = types.StringNull()
	}
	// Preserve api_key from plan (write-only, not returned by API).
	state.APIKey = plan.APIKey

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *llmConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state llmConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteLLMConnection(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting LLM connection",
			fmt.Sprintf("Unable to delete LLM connection %s, got error: %s", state.ID.ValueString(), err),
		)
	}
}

func (r *llmConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by connection ID: we need to list connections to find by ID.
	// The API provides GetLLMConnectionByName; for import we use the ID stored in state.
	// Since we don't have a GetByID method, we require the user to provide the connection name as import ID.
	conn, err := r.client.GetLLMConnectionByName(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing LLM connection",
			fmt.Sprintf("Unable to read LLM connection %s, got error: %s", req.ID, err),
		)
		return
	}

	state := llmConnectionResourceModel{
		ID:                types.StringValue(conn.ID),
		Name:              types.StringValue(conn.Name),
		Provider:          types.StringValue(conn.Provider),
		WithDefaultModels: types.BoolValue(conn.WithDefaultModels),
		// api_key cannot be imported.
		APIKey: types.StringNull(),
	}

	if conn.BaseURL != nil {
		state.BaseURL = types.StringValue(*conn.BaseURL)
	} else {
		state.BaseURL = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
