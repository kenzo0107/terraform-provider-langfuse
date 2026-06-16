package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kenzo0107/terraform-provider-langfuse/langfuse"
)

var _ provider.Provider = &langfuseProvider{}

type langfuseProvider struct {
	version string
}

type langfuseProviderModel struct {
	PublicKey types.String `tfsdk:"public_key"`
	SecretKey types.String `tfsdk:"secret_key"`
	Host      types.String `tfsdk:"host"`
}

func (p *langfuseProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "langfuse"
	resp.Version = p.version
}

func (p *langfuseProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Langfuse provider allows managing Langfuse resources via the Langfuse API.",
		Attributes: map[string]schema.Attribute{
			"public_key": schema.StringAttribute{
				MarkdownDescription: "Langfuse public API key. May also be provided via `LANGFUSE_PUBLIC_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"secret_key": schema.StringAttribute{
				MarkdownDescription: "Langfuse secret API key. May also be provided via `LANGFUSE_SECRET_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "Langfuse API host. May also be provided via `LANGFUSE_HOST` environment variable. Defaults to `https://cloud.langfuse.com`.",
				Optional:            true,
			},
		},
	}
}

func (p *langfuseProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	publicKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
	secretKey := os.Getenv("LANGFUSE_SECRET_KEY")
	host := os.Getenv("LANGFUSE_HOST")

	var config langfuseProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.PublicKey.IsNull() {
		publicKey = config.PublicKey.ValueString()
	}

	if !config.SecretKey.IsNull() {
		secretKey = config.SecretKey.ValueString()
	}

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if config.PublicKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("public_key"),
			"Unknown Langfuse Public Key",
			"The provider cannot create the Langfuse API client as there is an unknown configuration value for the Langfuse public key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the LANGFUSE_PUBLIC_KEY environment variable.",
		)
	}

	if config.SecretKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("secret_key"),
			"Unknown Langfuse Secret Key",
			"The provider cannot create the Langfuse API client as there is an unknown configuration value for the Langfuse secret key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the LANGFUSE_SECRET_KEY environment variable.",
		)
	}

	if publicKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("public_key"),
			"Missing Langfuse Public Key",
			"The provider cannot create the Langfuse API client as there is a missing or empty value for the Langfuse public key. "+
				"Set the public_key value in the configuration or use the LANGFUSE_PUBLIC_KEY environment variable.",
		)
	}

	if secretKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("secret_key"),
			"Missing Langfuse Secret Key",
			"The provider cannot create the Langfuse API client as there is a missing or empty value for the Langfuse secret key. "+
				"Set the secret_key value in the configuration or use the LANGFUSE_SECRET_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	opts := []langfuse.Option{}
	if host != "" {
		opts = append(opts, langfuse.WithHost(host))
	}

	client := langfuse.New(publicKey, secretKey, opts...)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *langfuseProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newProjectResource,
		newProjectMemberResource,
		newScoreConfigResource,
	}
}

func (p *langfuseProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newProjectDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &langfuseProvider{
			version: version,
		}
	}
}
