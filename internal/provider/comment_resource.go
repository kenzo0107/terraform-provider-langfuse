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

var _ resource.Resource = &commentResource{}
var _ resource.ResourceWithImportState = &commentResource{}

func newCommentResource() resource.Resource {
	return &commentResource{}
}

type commentResource struct {
	client *langfuse.Client
}

type commentResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ProjectID    types.String `tfsdk:"project_id"`
	ObjectType   types.String `tfsdk:"object_type"`
	ObjectID     types.String `tfsdk:"object_id"`
	Content      types.String `tfsdk:"content"`
	AuthorUserID types.String `tfsdk:"author_user_id"`
}

func (r *commentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_comment"
}

func (r *commentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Langfuse comment on a trace, observation, session, or prompt.\n\n> **Note:** The Langfuse API does not support updating or deleting comments. Destroying this resource only removes it from Terraform state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the comment.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The project ID the comment belongs to (read-only, set by Langfuse).",
				Computed:            true,
			},
			"object_type": schema.StringAttribute{
				MarkdownDescription: "The type of object the comment is attached to. Must be `TRACE`, `OBSERVATION`, `SESSION`, or `PROMPT`. Changing this creates a new comment.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the object the comment is attached to. Changing this creates a new comment.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The text content of the comment. Changing this creates a new comment.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"author_user_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the author user. Changing this creates a new comment.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *commentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *commentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan commentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &langfuse.CreateCommentRequest{
		ObjectType: plan.ObjectType.ValueString(),
		ObjectID:   plan.ObjectID.ValueString(),
		Content:    plan.Content.ValueString(),
	}
	if !plan.AuthorUserID.IsNull() && !plan.AuthorUserID.IsUnknown() {
		v := plan.AuthorUserID.ValueString()
		createReq.AuthorUserID = &v
	}

	commentID, err := r.client.CreateComment(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating comment",
			fmt.Sprintf("Unable to create comment, got error: %s", err),
		)
		return
	}

	cm, err := r.client.GetComment(ctx, commentID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading comment after creation",
			fmt.Sprintf("Unable to read comment %s after creation, got error: %s", commentID, err),
		)
		return
	}

	plan.ID = types.StringValue(cm.ID)
	plan.ProjectID = types.StringValue(cm.ProjectID)
	plan.ObjectType = types.StringValue(cm.ObjectType)
	plan.ObjectID = types.StringValue(cm.ObjectID)
	plan.Content = types.StringValue(cm.Content)
	if cm.AuthorUserID != nil {
		plan.AuthorUserID = types.StringValue(*cm.AuthorUserID)
	} else {
		plan.AuthorUserID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *commentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state commentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cm, err := r.client.GetComment(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading comment",
			fmt.Sprintf("Unable to read comment %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.ProjectID = types.StringValue(cm.ProjectID)
	state.ObjectType = types.StringValue(cm.ObjectType)
	state.ObjectID = types.StringValue(cm.ObjectID)
	state.Content = types.StringValue(cm.Content)
	if cm.AuthorUserID != nil {
		state.AuthorUserID = types.StringValue(*cm.AuthorUserID)
	} else {
		state.AuthorUserID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *commentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Updating comment",
		"In-place updates are not supported for langfuse_comment; all attributes require replacement.",
	)
}

func (r *commentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Langfuse has no delete API for comments; removing from state only.
}

func (r *commentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	cm, err := r.client.GetComment(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing comment",
			fmt.Sprintf("Unable to read comment %s, got error: %s", req.ID, err),
		)
		return
	}

	state := commentResourceModel{
		ID:        types.StringValue(cm.ID),
		ProjectID: types.StringValue(cm.ProjectID),
		ObjectType: types.StringValue(cm.ObjectType),
		ObjectID:  types.StringValue(cm.ObjectID),
		Content:   types.StringValue(cm.Content),
	}
	if cm.AuthorUserID != nil {
		state.AuthorUserID = types.StringValue(*cm.AuthorUserID)
	} else {
		state.AuthorUserID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
