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

var _ resource.Resource = &projectMemberResource{}
var _ resource.ResourceWithImportState = &projectMemberResource{}

func newProjectMemberResource() resource.Resource {
	return &projectMemberResource{}
}

type projectMemberResource struct {
	client *langfuse.Client
}

type projectMemberResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	UserID    types.String `tfsdk:"user_id"`
	Role      types.String `tfsdk:"role"`
	Email     types.String `tfsdk:"email"`
	Name      types.String `tfsdk:"name"`
}

func (r *projectMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_member"
}

func (r *projectMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a member's role within a Langfuse project. The user must already be a member of the organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the membership (format: `{project_id}/{user_id}`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user to add to the project.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role to assign to the user. Must be one of `OWNER`, `ADMIN`, `MEMBER`, or `VIEWER`.",
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email address of the member (read-only, populated after creation).",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The display name of the member (read-only, populated after creation).",
				Computed:            true,
			},
		},
	}
}

func (r *projectMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.UpsertProjectMembership(
		ctx,
		plan.ProjectID.ValueString(),
		plan.UserID.ValueString(),
		langfuse.MembershipRole(plan.Role.ValueString()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating project member",
			fmt.Sprintf("Unable to add member to project, got error: %s", err),
		)
		return
	}

	plan.ID = types.StringValue(plan.ProjectID.ValueString() + "/" + m.UserID)
	plan.UserID = types.StringValue(m.UserID)
	plan.Role = types.StringValue(string(m.Role))
	plan.Email = types.StringValue(m.Email)
	plan.Name = types.StringValue(m.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.GetProjectMembership(ctx, state.ProjectID.ValueString(), state.UserID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading project member",
			fmt.Sprintf("Unable to read membership for user %s in project %s, got error: %s", state.UserID.ValueString(), state.ProjectID.ValueString(), err),
		)
		return
	}

	state.Role = types.StringValue(string(m.Role))
	state.Email = types.StringValue(m.Email)
	state.Name = types.StringValue(m.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state projectMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.UpsertProjectMembership(
		ctx,
		state.ProjectID.ValueString(),
		state.UserID.ValueString(),
		langfuse.MembershipRole(plan.Role.ValueString()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating project member",
			fmt.Sprintf("Unable to update role for user %s in project %s, got error: %s", state.UserID.ValueString(), state.ProjectID.ValueString(), err),
		)
		return
	}

	state.Role = types.StringValue(string(m.Role))
	state.Email = types.StringValue(m.Email)
	state.Name = types.StringValue(m.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteProjectMembership(ctx, state.ProjectID.ValueString(), state.UserID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting project member",
			fmt.Sprintf("Unable to remove user %s from project %s, got error: %s", state.UserID.ValueString(), state.ProjectID.ValueString(), err),
		)
	}
}

func (r *projectMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format `{project_id}/{user_id}`.",
		)
		return
	}

	projectID, userID := parts[0], parts[1]

	m, err := r.client.GetProjectMembership(ctx, projectID, userID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing project member",
			fmt.Sprintf("Unable to read membership for user %s in project %s, got error: %s", userID, projectID, err),
		)
		return
	}

	state := projectMemberResourceModel{
		ID:        types.StringValue(projectID + "/" + userID),
		ProjectID: types.StringValue(projectID),
		UserID:    types.StringValue(userID),
		Role:      types.StringValue(string(m.Role)),
		Email:     types.StringValue(m.Email),
		Name:      types.StringValue(m.Name),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
