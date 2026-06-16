package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kenzo0107/terraform-provider-langfuse/langfuse"
)

var _ resource.Resource = &blobStorageIntegrationResource{}
var _ resource.ResourceWithImportState = &blobStorageIntegrationResource{}

func newBlobStorageIntegrationResource() resource.Resource {
	return &blobStorageIntegrationResource{}
}

type blobStorageIntegrationResource struct {
	client *langfuse.Client
}

type blobStorageIntegrationResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Type            types.String `tfsdk:"type"`
	BucketName      types.String `tfsdk:"bucket_name"`
	Prefix          types.String `tfsdk:"prefix"`
	Region          types.String `tfsdk:"region"`
	Endpoint        types.String `tfsdk:"endpoint"`
	ExportPrefix    types.String `tfsdk:"export_prefix"`
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	Enabled         types.Bool   `tfsdk:"enabled"`
}

func (r *blobStorageIntegrationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blob_storage_integration"
}

func (r *blobStorageIntegrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Langfuse blob storage integration for exporting data. " +
			"Note: `access_key_id` and `secret_access_key` are write-only and cannot be imported.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the blob storage integration.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The storage backend type. Must be one of `S3`, `AZURE_BLOB`, or `GCS`.",
				Required:            true,
			},
			"bucket_name": schema.StringAttribute{
				MarkdownDescription: "The name of the storage bucket.",
				Required:            true,
			},
			"prefix": schema.StringAttribute{
				MarkdownDescription: "An optional prefix for objects stored in the bucket.",
				Optional:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The region of the storage bucket (for S3-compatible storage).",
				Optional:            true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "A custom endpoint URL (for S3-compatible storage or self-hosted).",
				Optional:            true,
			},
			"export_prefix": schema.StringAttribute{
				MarkdownDescription: "An optional prefix used when exporting data.",
				Optional:            true,
			},
			"access_key_id": schema.StringAttribute{
				MarkdownDescription: "The access key ID for authenticating with the storage backend. " +
					"This is write-only and will not be read back from the API.",
				Optional:  true,
				Sensitive: true,
			},
			"secret_access_key": schema.StringAttribute{
				MarkdownDescription: "The secret access key for authenticating with the storage backend. " +
					"This is write-only and will not be read back from the API.",
				Optional:  true,
				Sensitive: true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the integration is enabled. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *blobStorageIntegrationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildUpsertBlobStorageRequest(plan *blobStorageIntegrationResourceModel) *langfuse.UpsertBlobStorageIntegrationRequest {
	upsertReq := &langfuse.UpsertBlobStorageIntegrationRequest{
		Type:       plan.Type.ValueString(),
		BucketName: plan.BucketName.ValueString(),
	}

	if !plan.Prefix.IsNull() && !plan.Prefix.IsUnknown() {
		v := plan.Prefix.ValueString()
		upsertReq.Prefix = &v
	}
	if !plan.Region.IsNull() && !plan.Region.IsUnknown() {
		v := plan.Region.ValueString()
		upsertReq.Region = &v
	}
	if !plan.Endpoint.IsNull() && !plan.Endpoint.IsUnknown() {
		v := plan.Endpoint.ValueString()
		upsertReq.Endpoint = &v
	}
	if !plan.ExportPrefix.IsNull() && !plan.ExportPrefix.IsUnknown() {
		v := plan.ExportPrefix.ValueString()
		upsertReq.ExportPrefix = &v
	}
	if !plan.AccessKeyID.IsNull() && !plan.AccessKeyID.IsUnknown() {
		v := plan.AccessKeyID.ValueString()
		upsertReq.AccessKeyID = &v
	}
	if !plan.SecretAccessKey.IsNull() && !plan.SecretAccessKey.IsUnknown() {
		v := plan.SecretAccessKey.ValueString()
		upsertReq.SecretAccessKey = &v
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		v := plan.Enabled.ValueBool()
		upsertReq.Enabled = &v
	}

	return upsertReq
}

func applyBlobStorageIntegrationToModel(b *langfuse.BlobStorageIntegration, state *blobStorageIntegrationResourceModel) {
	state.ID = types.StringValue(b.ID)
	state.Type = types.StringValue(b.Type)
	state.BucketName = types.StringValue(b.BucketName)
	state.Enabled = types.BoolValue(b.Enabled)

	if b.Prefix != nil {
		state.Prefix = types.StringValue(*b.Prefix)
	} else {
		state.Prefix = types.StringNull()
	}

	if b.Region != nil {
		state.Region = types.StringValue(*b.Region)
	} else {
		state.Region = types.StringNull()
	}

	if b.Endpoint != nil {
		state.Endpoint = types.StringValue(*b.Endpoint)
	} else {
		state.Endpoint = types.StringNull()
	}

	if b.ExportPrefix != nil {
		state.ExportPrefix = types.StringValue(*b.ExportPrefix)
	} else {
		state.ExportPrefix = types.StringNull()
	}
	// access_key_id and secret_access_key are not returned by the API; preserve state values.
}

func (r *blobStorageIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan blobStorageIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	b, err := r.client.UpsertBlobStorageIntegration(ctx, buildUpsertBlobStorageRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Creating blob storage integration",
			fmt.Sprintf("Unable to create blob storage integration, got error: %s", err),
		)
		return
	}

	applyBlobStorageIntegrationToModel(b, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *blobStorageIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state blobStorageIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	b, err := r.client.GetBlobStorageIntegration(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading blob storage integration",
			fmt.Sprintf("Unable to read blob storage integration %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	applyBlobStorageIntegrationToModel(b, &state)
	// Credentials are preserved from prior state (not returned by the API).

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *blobStorageIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state blobStorageIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	b, err := r.client.UpsertBlobStorageIntegration(ctx, buildUpsertBlobStorageRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating blob storage integration",
			fmt.Sprintf("Unable to update blob storage integration %s, got error: %s", state.ID.ValueString(), err),
		)
		return
	}

	applyBlobStorageIntegrationToModel(b, &state)
	// Preserve credentials from plan (write-only, not returned by API).
	state.AccessKeyID = plan.AccessKeyID
	state.SecretAccessKey = plan.SecretAccessKey

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *blobStorageIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state blobStorageIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteBlobStorageIntegration(ctx, state.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*langfuse.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Deleting blob storage integration",
			fmt.Sprintf("Unable to delete blob storage integration %s, got error: %s", state.ID.ValueString(), err),
		)
	}
}

func (r *blobStorageIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	b, err := r.client.GetBlobStorageIntegration(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing blob storage integration",
			fmt.Sprintf("Unable to read blob storage integration %s, got error: %s", req.ID, err),
		)
		return
	}

	var state blobStorageIntegrationResourceModel
	applyBlobStorageIntegrationToModel(b, &state)
	// access_key_id and secret_access_key cannot be recovered after import.
	state.AccessKeyID = types.StringNull()
	state.SecretAccessKey = types.StringNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
