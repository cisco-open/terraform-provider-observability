// Copyright (c) HashiCorp, Inc.type
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"terraform-provider-cop/internal/api"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure COPProvider satisfies various provider interfaces.
var _ provider.Provider = &COPProvider{}
var _ provider.ProviderWithFunctions = &COPProvider{}

// COPProvider defines the provider implementation.
type COPProvider struct {
	version string
}

// COPProviderModel describes the provider data model.
type COPProviderModel struct {
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	URL        types.String `tfsdk:"url"`
	AuthMethod types.String `tfsdk:"auth_method"`
	Tenant     types.String `tfsdk:"tenant"`
}

func (p *COPProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cop"
	resp.Version = p.version
}

func (p *COPProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"auth_method": schema.StringAttribute{
				MarkdownDescription: "Authentication type selected for COP API requests. Possible values(oauth, headless)",
				Required:            true,
			},
			"tenant": schema.StringAttribute{
				MarkdownDescription: "Tenant ID used to make requests to API",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username to authenticate using headless",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password to authenticate using headless",
				Optional:            true,
				Sensitive:           true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "URL used when authentication eg. https://mytenant.com",
				Optional:            true,
			},
		},
	}
}

func (p *COPProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data COPProviderModel

	// Retrieve provider data from configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if data.AuthMethod.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("auth_method"),
			"Unknown cop API auth_method",
			"Please make sure you configure the auth_method field",
		)
	}

	if data.Tenant.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("tenant"),
			"Unknown cop API tenant",
			"Please make sure you configure the tenant field",
		)
	}

	if data.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown cop API username",
			"Please make sure you configure the username field",
		)
	}

	if data.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown cop API password",
			"Please make sure you configure the password field",
		)
	}

	if data.URL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Unknown cop API url",
			"Please make sure you configure the url field",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	username := os.Getenv("COP_USERNAME")
	password := os.Getenv("COP_PASSWORD")
	authMethod := os.Getenv("COP_AUTH_METHOD")
	tenantID := os.Getenv("COP_TENANT")
	url := os.Getenv("URL")

	tflog.Debug(ctx, fmt.Sprintf("Terraform username is %s", data.Username))
	tflog.Debug(ctx, fmt.Sprintf("Terraform password is %s", data.Password))
	tflog.Debug(ctx, fmt.Sprintf("Terraform url is %s", data.URL))
	tflog.Debug(ctx, fmt.Sprintf("Terraform tenant is %s", data.Tenant))

	tflog.Debug(ctx, fmt.Sprintf("Terraform auth_method is %s", data.AuthMethod))
	tflog.Debug(ctx, fmt.Sprintf("Terraform auth_method FROM ENV is %s", authMethod))

	if !data.Username.IsNull() {
		username = data.Username.ValueString()
	}

	if !data.Password.IsNull() {
		password = data.Password.ValueString()
	}

	if !data.URL.IsNull() {
		url = data.URL.ValueString()
	}

	if !data.Tenant.IsNull() {
		tenantID = data.Tenant.ValueString()
	}

	if !data.AuthMethod.IsNull() {
		authMethod = data.AuthMethod.ValueString()
	}

	if authMethod == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("auth_method"),
			"Missing cop API auth_method",
			"SET the COP_AUTH_METHOD env var or the config",
		)
		tflog.Error(ctx, "Missing or empty value for auth_method attribute")
	}

	switch authMethod {
	case "oauth":
		if url == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("url"),
				"Missing cop API url",
				"SET the COP_URL env var or the config",
			)
		}

		if tenantID == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("tenant"),
				"Missing cop API tenant",
				"SET the COP_TENANT env var or the config",
			)
		}
	case "headless":
		if username == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("username"),
				"Missing cop API username",
				"SET the COP_USERNAME env var or the config",
			)
		}

		if password == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("password"),
				"Missing cop API password",
				"SET the COP_PASSWORD env var or the config",
			)
		}
	}

	// exit if any of the required attributes is missing
	// based on our current auth_method
	if resp.Diagnostics.HasError() {
		return
	}

	appdClient := &api.AppdClient{
		AuthMethod: authMethod,
		Username:   username,
		Password:   password,
		URL:        url,
		Tenant:     tenantID,
		APIClient:  http.DefaultClient,
	}

	err := appdClient.Login()
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Failed to authenticate to COP client: %s", err.Error()))
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfull authentication to COP client using %s", appdClient.AuthMethod))

	// TODO change this to a real client
	resp.DataSourceData = appdClient
	resp.ResourceData = appdClient
}

func (p *COPProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *COPProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewTypeDataSource,
	}
}

func (p *COPProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewExampleFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &COPProvider{
			version: version,
		}
	}
}
