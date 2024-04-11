// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/cisco-open/terraform-provider-observability/internal/api"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ObjectResource{}
var _ resource.ResourceWithImportState = &ObjectResource{}

func NewObjectResource() resource.Resource {
	return &ObjectResource{}
}

// ObjectResource defines the resource implementation.
type ObjectResource struct {
	client *api.AppdClient
}

// ObjectResourceModel describes the resource data model.
type ObjectResourceModel struct {
	Typename   types.String  `tfsdk:"type_name"`
	ObjectID   types.String  `tfsdk:"object_id"`
	TenantID   types.String  `tfsdk:"layer_id"`
	TenantType types.String  `tfsdk:"layer_type"`
	Data       types.Dynamic `tfsdk:"data"`
	ID         types.String  `tfsdk:"id"`
}

func (r *ObjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object"
}

func (r *ObjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Object resource",

		Attributes: map[string]schema.Attribute{
			"type_name": schema.StringAttribute{
				MarkdownDescription: "Specifies the fully qualified type name used to get the type",
				Required:            true,
			},
			"object_id": schema.StringAttribute{
				MarkdownDescription: "Spepcified the object ID for the particular object to get",
				Required:            true,
			},
			"layer_id": schema.StringAttribute{
				MarkdownDescription: "Specifies the layer ID where the object resides",
				Required:            true,
			},
			"layer_type": schema.StringAttribute{
				MarkdownDescription: "Specifies the layer type where the object resides",
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

func (r *ObjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.AppdClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.AppdClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

//nolint:gocritic // Terraform framework requires the method signature to be as is
func (r *ObjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ObjectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
	//     return
	// }

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.ID = types.StringValue("example-id")

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:gocritic // Terraform framework requires the method signature to be as is
func (r *ObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ObjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// issue the API call
	typeName := data.Typename.ValueString()
	objID := data.ObjectID.ValueString()
	layerID := data.TenantID.ValueString()
	layerType := data.TenantType.ValueString()

	result, err := r.client.GetObject(typeName, objID, layerID, layerType)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read object of type %s with id %s", typeName, objID),
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("\n\nResponse is %+v\n\n", string(result)))

	data.Data = types.DynamicValue(types.StringValue(string(result)))
	tflog.Trace(ctx, "read a resource")

	// set the placeholder value for testing purposses
	data.ID = types.StringValue("placeholder")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:gocritic // Terraform framework requires the method signature to be as is
func (r *ObjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ObjectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:gocritic // Terraform framework requires the method signature to be as is
func (r *ObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ObjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ObjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
