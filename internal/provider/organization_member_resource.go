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

var _ resource.Resource = &organizationMemberResource{}
var _ resource.ResourceWithImportState = &organizationMemberResource{}

func newOrganizationMemberResource() resource.Resource {
	return &organizationMemberResource{}
}

type organizationMemberResource struct {
	client *langfuse.Client
}

type organizationMemberResourceModel struct {
	ID     types.String `tfsdk:"id"`
	UserID types.String `tfsdk:"user_id"`
	Role   types.String `tfsdk:"role"`
	Email  types.String `tfsdk:"email"`
	Name   types.String `tfsdk:"name"`
}

func (r *organizationMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_member"
}

func (r *organizationMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a member's role within the Langfuse organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the membership (set to `user_id`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user to add to the organization.",
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

func (r *organizationMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.UpsertOrganizationMembership(
		ctx,
		plan.UserID.ValueString(),
		langfuse.MembershipRole(plan.Role.ValueString()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating organization member",
			fmt.Sprintf("Unable to add member to organization, got error: %s", err),
		)
		return
	}

	plan.ID = types.StringValue(m.UserID)
	plan.UserID = types.StringValue(m.UserID)
	plan.Role = types.StringValue(string(m.Role))
	plan.Email = types.StringValue(m.Email)
	plan.Name = types.StringValue(m.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *organizationMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.GetOrganizationMembership(ctx, state.UserID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading organization member",
			fmt.Sprintf("Unable to read organization membership for user %s, got error: %s", state.UserID.ValueString(), err),
		)
		return
	}

	state.Role = types.StringValue(string(m.Role))
	state.Email = types.StringValue(m.Email)
	state.Name = types.StringValue(m.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state organizationMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.UpsertOrganizationMembership(
		ctx,
		state.UserID.ValueString(),
		langfuse.MembershipRole(plan.Role.ValueString()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating organization member",
			fmt.Sprintf("Unable to update role for user %s in organization, got error: %s", state.UserID.ValueString(), err),
		)
		return
	}

	state.Role = types.StringValue(string(m.Role))
	state.Email = types.StringValue(m.Email)
	state.Name = types.StringValue(m.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteOrganizationMembership(ctx, state.UserID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting organization member",
			fmt.Sprintf("Unable to remove user %s from organization, got error: %s", state.UserID.ValueString(), err),
		)
	}
}

func (r *organizationMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	userID := req.ID

	m, err := r.client.GetOrganizationMembership(ctx, userID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing organization member",
			fmt.Sprintf("Unable to read organization membership for user %s, got error: %s", userID, err),
		)
		return
	}

	state := organizationMemberResourceModel{
		ID:     types.StringValue(userID),
		UserID: types.StringValue(userID),
		Role:   types.StringValue(string(m.Role)),
		Email:  types.StringValue(m.Email),
		Name:   types.StringValue(m.Name),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
