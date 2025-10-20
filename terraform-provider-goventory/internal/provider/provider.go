package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// provider implementation.
type goventoryProvider struct {
	version string
}

// maps provider schema data to a Go type
type goventoryProviderModel struct {
	ServerURL types.String `tfsdk:"server_url"`
}

// new GoVentory provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &goventoryProvider{
			version: version,
		}
	}
}

// provider's metadata
func (p *goventoryProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "goventory"
	resp.Version = p.version
}

// provider's schema
func (p *goventoryProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "GoVentory provider to manage Ansible inventory hosts.",
		Attributes: map[string]schema.Attribute{
			"server_url": schema.StringAttribute{
				Description: "The URL of the GoVentory API server.",
				Required:    true,
			},
		},
	}
}

// configures the provider
func (p *goventoryProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config goventoryProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverURL := config.ServerURL.ValueString()

	client := &http.Client{}

	resp.DataSourceData = client
	resp.ResourceData = map[string]interface{}{
		"client":    client,
		"server_url": serverURL,
	}
}

// returns the provider's resources
func (p *goventoryProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewHostResource,
	}
}

// returns the provider's data sources
func (p *goventoryProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
