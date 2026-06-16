package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kenzo0107/terraform-provider-langfuse/langfuse"
)

var _ resource.Resource = &scimUserResource{}
var _ resource.ResourceWithImportState = &scimUserResource{}

func newSCIMUserResource() resource.Resource {
	return &scimUserResource{}
}

type scimUserResource struct {
	client *langfuse.Client
}

type scimUserResourceModel struct {
	ID          types.String `tfsdk:"id"`
	UserName    types.String `tfsdk:"user_name"`
	GivenName   types.String `tfsdk:"given_name"`
	FamilyName  types.String `tfsdk:"family_name"`
	Email       types.String `tfsdk:"email"`
	Active      types.Bool   `tfsdk:"active"`
	ExternalID  types.String `tfsdk:"external_id"`
	Password    types.String `tfsdk:"password"`
}

func (r *scimUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scim_user"
}

func (r *scimUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Langfuse SCIM user. Requires an organization-scoped API key with SCIM permissions.\n\n> **Note:** `password` is write-only and cannot be recovered via import.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the SCIM user.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_name": schema.StringAttribute{
				MarkdownDescription: "The username (typically an email address). Changing this creates a new user.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The primary email address of the user. Changing this creates a new user.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"given_name": schema.StringAttribute{
				MarkdownDescription: "The given (first) name of the user. Changing this creates a new user.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"family_name": schema.StringAttribute{
				MarkdownDescription: "The family (last) name of the user. Changing this creates a new user.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the user is active. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"external_id": schema.StringAttribute{
				MarkdownDescription: "An optional external identifier for the user. Changing this creates a new user.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The initial password for the user. Write-only; not read back from the API.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *scimUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *scimUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan scimUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &langfuse.CreateSCIMUserRequest{
		UserName: plan.UserName.ValueString(),
		Emails: []langfuse.SCIMEmail{
			{Value: plan.Email.ValueString(), Primary: true},
		},
		Name: langfuse.SCIMUserName{
			GivenName:  plan.GivenName.ValueString(),
			FamilyName: plan.FamilyName.ValueString(),
		},
	}
	if !plan.Active.IsNull() && !plan.Active.IsUnknown() {
		v := plan.Active.ValueBool()
		createReq.Active = &v
	}
	if !plan.ExternalID.IsNull() && !plan.ExternalID.IsUnknown() {
		v := plan.ExternalID.ValueString()
		createReq.ExternalID = &v
	}
	if !plan.Password.IsNull() && !plan.Password.IsUnknown() {
		v := plan.Password.ValueString()
		createReq.Password = &v
	}

	u, err := r.client.CreateSCIMUser(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating SCIM user",
			fmt.Sprintf("Unable to create SCIM user, got error: %s", err),
		)
		return
	}

	plan.ID = types.StringValue(u.ID)
	plan.UserName = types.StringValue(u.UserName)
	plan.Active = types.BoolValue(u.Active)
	plan.GivenName = types.StringValue(u.Name.GivenName)
	plan.FamilyName = types.StringValue(u.Name.FamilyName)
	if len(u.Emails) > 0 {
		plan.Email = types.StringValue(u.Emails[0].Value)
	}
	if u.ExternalID != nil {
		plan.ExternalID = types.StringValue(*u.ExternalID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scimUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scimUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	u, err := r.client.GetSCIMUser(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading SCIM user",
			fmt.Sprintf("Unable to read SCIM user %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.UserName = types.StringValue(u.UserName)
	state.Active = types.BoolValue(u.Active)
	state.GivenName = types.StringValue(u.Name.GivenName)
	state.FamilyName = types.StringValue(u.Name.FamilyName)
	if len(u.Emails) > 0 {
		state.Email = types.StringValue(u.Emails[0].Value)
	}
	if u.ExternalID != nil {
		state.ExternalID = types.StringValue(*u.ExternalID)
	} else {
		state.ExternalID = types.StringNull()
	}
	// password is write-only; preserve from state

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *scimUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Updating SCIM user",
		"In-place updates are not supported for langfuse_scim_user; all attributes require replacement.",
	)
}

func (r *scimUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state scimUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSCIMUser(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting SCIM user",
			fmt.Sprintf("Unable to delete SCIM user %s, got error: %s", state.ID.ValueString(), err),
		)
	}
}

func (r *scimUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	u, err := r.client.GetSCIMUser(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing SCIM user",
			fmt.Sprintf("Unable to read SCIM user %s, got error: %s", req.ID, err),
		)
		return
	}

	state := scimUserResourceModel{
		ID:       types.StringValue(u.ID),
		UserName: types.StringValue(u.UserName),
		Active:   types.BoolValue(u.Active),
		GivenName: types.StringValue(u.Name.GivenName),
		FamilyName: types.StringValue(u.Name.FamilyName),
		Password: types.StringNull(),
	}
	if len(u.Emails) > 0 {
		state.Email = types.StringValue(u.Emails[0].Value)
	} else {
		state.Email = types.StringNull()
	}
	if u.ExternalID != nil {
		state.ExternalID = types.StringValue(*u.ExternalID)
	} else {
		state.ExternalID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
