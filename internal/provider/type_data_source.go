// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"terraform-provider-cop/internal/api"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &TypeDataSource{}

func NewTypeDataSource() datasource.DataSource {
	return &TypeDataSource{}
}

// TypeDataSource defines the data source implementation.
type TypeDataSource struct {
	client *api.AppdClient
}

// TypeDataSourceModel describes the data source data model.
type TypeDataSourceModel struct {
	Typename types.String  `tfsdk:"type_name"`
	Data     types.Dynamic `tfsdk:"data"`
	Id       types.String  `tfsdk:"id"`
}

func (d *TypeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_type"
}

func (d *TypeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Type data source",
		Attributes: map[string]schema.Attribute{
			"type_name": schema.StringAttribute{
				MarkdownDescription: "Specifies the fully qualified type name used to get the type",
				Required:            true,
			},
			"data": schema.DynamicAttribute{
				MarkdownDescription: "JSON schema of the returned type",
				Optional:            true,
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Used to provide compatibility for testing framework",
				Computed:            true,
			},
		},
	}
}

func (d *TypeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.AppdClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.AppdClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *TypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TypeDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// issue the API call
	typeName := data.Typename
	result, err := d.client.GetType(typeName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read type %s", typeName),
			err.Error(),
		)
		return
	}

	// convert result to Terraform Schema and populate
	//response := make(map[string]any)
	//json.Unmarshal(result, &response)

	tflog.Debug(ctx, fmt.Sprintf("\n\nResponse is %+v\n\n", string(result)))

	data.Data = types.DynamicValue(types.StringValue(string(result)))
	tflog.Trace(ctx, "read a data source")

	// set the placeholder value for testing purposses
	data.Id = types.StringValue("placeholder")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
