package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kenzo0107/terraform-provider-langfuse/langfuse"
)

var _ resource.Resource = &promptResource{}
var _ resource.ResourceWithImportState = &promptResource{}

func newPromptResource() resource.Resource {
	return &promptResource{}
}

type promptResource struct {
	client *langfuse.Client
}

type promptResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Text     types.String `tfsdk:"text"`
	Messages types.List   `tfsdk:"messages"`
	Labels   types.List   `tfsdk:"labels"`
	Tags     types.List   `tfsdk:"tags"`
	Version  types.Int64  `tfsdk:"version"`
}

type chatMessageModel struct {
	Role    types.String `tfsdk:"role"`
	Content types.String `tfsdk:"content"`
}

var chatMessageAttrTypes = map[string]attr.Type{
	"role":    types.StringType,
	"content": types.StringType,
}

func (r *promptResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompt"
}

func (r *promptResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Langfuse prompt. Each update creates a new version of the prompt.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the prompt (format: `{name}:v{version}`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the prompt.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the prompt. Must be `text` or `chat`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"text": schema.StringAttribute{
				MarkdownDescription: "The prompt text content (for `type = text`).",
				Optional:            true,
			},
			"messages": schema.ListNestedAttribute{
				MarkdownDescription: "Chat messages (for `type = chat`).",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role": schema.StringAttribute{
							MarkdownDescription: "The role of the message author (e.g. `user`, `assistant`, `system`).",
							Required:            true,
						},
						"content": schema.StringAttribute{
							MarkdownDescription: "The content of the message.",
							Required:            true,
						},
					},
				},
			},
			"labels": schema.ListAttribute{
				MarkdownDescription: "Labels to attach to the prompt version.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags to attach to the prompt.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "The version number of the prompt (assigned by Langfuse after creation).",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *promptResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func chatMessagesFromList(ctx context.Context, list types.List) ([]langfuse.ChatMessage, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	var models []chatMessageModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, fmt.Errorf("reading messages from plan")
	}

	msgs := make([]langfuse.ChatMessage, len(models))
	for i, m := range models {
		msgs[i] = langfuse.ChatMessage{
			Role:    m.Role.ValueString(),
			Content: m.Content.ValueString(),
		}
	}
	return msgs, nil
}

func chatMessagesToList(msgs []langfuse.ChatMessage) (types.List, error) {
	if len(msgs) == 0 {
		return types.ListValueMust(types.ObjectType{AttrTypes: chatMessageAttrTypes}, []attr.Value{}), nil
	}

	elems := make([]attr.Value, len(msgs))
	for i, m := range msgs {
		obj, diags := types.ObjectValue(chatMessageAttrTypes, map[string]attr.Value{
			"role":    types.StringValue(m.Role),
			"content": types.StringValue(m.Content),
		})
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: chatMessageAttrTypes}), fmt.Errorf("building message object")
		}
		elems[i] = obj
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: chatMessageAttrTypes}, elems)
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: chatMessageAttrTypes}), fmt.Errorf("building messages list")
	}
	return list, nil
}

func stringSliceFromList(ctx context.Context, list types.List) ([]string, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	var values []string
	if diags := list.ElementsAs(ctx, &values, false); diags.HasError() {
		return nil, fmt.Errorf("reading list of strings")
	}
	return values, nil
}

func stringSliceToList(values []string) types.List {
	if len(values) == 0 {
		return types.ListValueMust(types.StringType, []attr.Value{})
	}

	elems := make([]attr.Value, len(values))
	for i, v := range values {
		elems[i] = types.StringValue(v)
	}
	return types.ListValueMust(types.StringType, elems)
}

func buildCreatePromptRequest(ctx context.Context, plan *promptResourceModel) (*langfuse.CreatePromptRequest, error) {
	createReq := &langfuse.CreatePromptRequest{
		Name: plan.Name.ValueString(),
		Type: plan.Type.ValueString(),
	}

	if !plan.Text.IsNull() && !plan.Text.IsUnknown() {
		createReq.Text = plan.Text.ValueString()
	}

	msgs, err := chatMessagesFromList(ctx, plan.Messages)
	if err != nil {
		return nil, err
	}
	createReq.Messages = msgs

	labels, err := stringSliceFromList(ctx, plan.Labels)
	if err != nil {
		return nil, err
	}
	createReq.Labels = labels

	tags, err := stringSliceFromList(ctx, plan.Tags)
	if err != nil {
		return nil, err
	}
	createReq.Tags = tags

	return createReq, nil
}

func applyPromptResponseToModel(_ context.Context, p *langfuse.PromptResponse, state *promptResourceModel) error {
	state.ID = types.StringValue(fmt.Sprintf("%s:v%d", p.Name, p.Version))
	state.Name = types.StringValue(p.Name)
	state.Type = types.StringValue(p.Type)
	state.Version = types.Int64Value(int64(p.Version))
	state.Text = types.StringValue(p.TextContent)

	msgs, err := chatMessagesToList(p.Messages)
	if err != nil {
		return err
	}
	state.Messages = msgs
	state.Labels = stringSliceToList(p.Labels)
	state.Tags = stringSliceToList(p.Tags)

	return nil
}

func (r *promptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan promptResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, err := buildCreatePromptRequest(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Creating prompt", err.Error())
		return
	}

	p, err := r.client.CreatePrompt(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating prompt",
			fmt.Sprintf("Unable to create prompt, got error: %s", err),
		)
		return
	}

	if err := applyPromptResponseToModel(ctx, p, &plan); err != nil {
		resp.Diagnostics.AddError("Creating prompt", fmt.Sprintf("Unable to map response: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *promptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state promptResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := r.client.GetPrompt(ctx, state.Name.ValueString(), int(state.Version.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading prompt",
			fmt.Sprintf("Unable to read prompt %s, got error: %s", state.Name.ValueString(), err),
		)
		return
	}

	if err := applyPromptResponseToModel(ctx, p, &state); err != nil {
		resp.Diagnostics.AddError("Reading prompt", fmt.Sprintf("Unable to map response: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *promptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state promptResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Updating creates a new version of the prompt.
	createReq, err := buildCreatePromptRequest(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Updating prompt", err.Error())
		return
	}

	p, err := r.client.CreatePrompt(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating prompt",
			fmt.Sprintf("Unable to create new version for prompt %s, got error: %s", state.Name.ValueString(), err),
		)
		return
	}

	if err := applyPromptResponseToModel(ctx, p, &state); err != nil {
		resp.Diagnostics.AddError("Updating prompt", fmt.Sprintf("Unable to map response: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *promptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state promptResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePrompt(ctx, state.Name.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting prompt",
			fmt.Sprintf("Unable to delete prompt %s, got error: %s", state.Name.ValueString(), err),
		)
	}
}

func (r *promptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by name; reads the latest version (version 0 signals latest).
	p, err := r.client.GetPrompt(ctx, req.ID, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing prompt",
			fmt.Sprintf("Unable to read prompt %s, got error: %s", req.ID, err),
		)
		return
	}

	var state promptResourceModel
	if err := applyPromptResponseToModel(ctx, p, &state); err != nil {
		resp.Diagnostics.AddError("Importing prompt", fmt.Sprintf("Unable to map response: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
