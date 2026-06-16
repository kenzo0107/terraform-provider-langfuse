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

var _ resource.Resource = &projectAPIKeyResource{}

func newProjectAPIKeyResource() resource.Resource {
	return &projectAPIKeyResource{}
}

type projectAPIKeyResource struct {
	client *langfuse.Client
}

type projectAPIKeyResourceModel struct {
	ID               types.String `tfsdk:"id"`
	ProjectID        types.String `tfsdk:"project_id"`
	Note             types.String `tfsdk:"note"`
	PublicKey        types.String `tfsdk:"public_key"`
	SecretKey        types.String `tfsdk:"secret_key"`
	DisplaySecretKey types.String `tfsdk:"display_secret_key"`
}

func (r *projectAPIKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_api_key"
}

func (r *projectAPIKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Langfuse project API key. " +
			"The `secret_key` is only available at creation time and is stored in Terraform state. " +
			"Import is not supported because the secret key cannot be recovered after the initial creation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the API key.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project this API key belongs to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"note": schema.StringAttribute{
				MarkdownDescription: "An optional note to identify the API key.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "The public key (read-only, assigned by Langfuse).",
				Computed:            true,
			},
			"secret_key": schema.StringAttribute{
				MarkdownDescription: "The secret key. Only available immediately after creation; stored in Terraform state. " +
					"This value cannot be recovered if the state is lost.",
				Computed:  true,
				Sensitive: true,
			},
			"display_secret_key": schema.StringAttribute{
				MarkdownDescription: "A partially masked representation of the secret key.",
				Computed:            true,
			},
		},
	}
}

func (r *projectAPIKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectAPIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectAPIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var note *string
	if !plan.Note.IsNull() && !plan.Note.IsUnknown() {
		v := plan.Note.ValueString()
		note = &v
	}

	key, err := r.client.CreateProjectAPIKey(ctx, plan.ProjectID.ValueString(), note)
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating project API key",
			fmt.Sprintf("Unable to create project API key, got error: %s", err),
		)
		return
	}

	plan.ID = types.StringValue(key.ID)
	plan.PublicKey = types.StringValue(key.PublicKey)
	plan.SecretKey = types.StringValue(key.SecretKey)
	plan.DisplaySecretKey = types.StringValue(key.DisplaySecretKey)

	if key.Note != nil {
		plan.Note = types.StringValue(*key.Note)
	} else {
		plan.Note = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keys, err := r.client.GetProjectAPIKeys(ctx, state.ProjectID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading project API key",
			fmt.Sprintf("Unable to list project API keys for project %s, got error: %s", state.ProjectID.ValueString(), err),
		)
		return
	}

	var found *langfuse.APIKey
	for i := range keys {
		if keys[i].ID == state.ID.ValueString() {
			found = &keys[i]
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.PublicKey = types.StringValue(found.PublicKey)
	state.DisplaySecretKey = types.StringValue(found.DisplaySecretKey)

	if found.Note != nil {
		state.Note = types.StringValue(*found.Note)
	} else {
		state.Note = types.StringNull()
	}
	// secret_key is not returned by the list API; preserve whatever is stored in state.

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectAPIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All mutable attributes have RequiresReplace, so Update should never be called.
	resp.Diagnostics.AddError(
		"Updating project API key",
		"Project API key does not support in-place updates. All changes require replacement.",
	)
}

func (r *projectAPIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteProjectAPIKey(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting project API key",
			fmt.Sprintf("Unable to delete project API key %s, got error: %s", state.ID.ValueString(), err),
		)
	}
}
