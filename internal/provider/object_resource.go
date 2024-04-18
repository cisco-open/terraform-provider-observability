// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cisco-open/terraform-provider-observability/internal/api"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	TypeName  types.String `tfsdk:"type_name"`
	ObjectID  types.String `tfsdk:"object_id"`
	LayerID   types.String `tfsdk:"layer_id"`
	LayerType types.String `tfsdk:"layer_type"`
	Data      types.String `tfsdk:"data"`
	ID        types.String `tfsdk:"id"`
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
				Optional:            true,
			},
			"layer_id": schema.StringAttribute{
				MarkdownDescription: "Specifies the layer ID where the object resides",
				Required:            true,
			},
			"layer_type": schema.StringAttribute{
				MarkdownDescription: "Specifies the layer type where the object resides",
				Required:            true,
			},
			"data": schema.StringAttribute{
				MarkdownDescription: "JSON schema of the returned object",
				Optional:            true,
				Validators: []validator.String{
					IsValidJsonString{},
				},
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
	tflog.Debug(ctx, "Create method invoked")
	var data ObjectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// issue the API call
	typeName := data.TypeName.ValueString()
	layerType := data.LayerType.ValueString()
	layerID := data.LayerID.ValueString()
	jsonPayload := []byte(data.Data.ValueString())

	err := r.client.CreateObject(typeName, layerID, layerType, jsonPayload)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read type %s", typeName),
			err.Error(),
		)
		return
	}

	// set the placeholder value for testing purposses
	data.ID = types.StringValue("placeholder")

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:gocritic // Terraform framework requires the method signature to be as is
func (r *ObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read method invoked")
	var data ObjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// issue the API call
	typeName := data.TypeName.ValueString()
	objID := data.ObjectID.ValueString()
	layerID := data.LayerID.ValueString()
	layerType := data.LayerType.ValueString()
	currentDataPayload := data.Data.ValueString()

	result, err := r.client.GetObject(typeName, objID, layerID, layerType)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read object of type %s with id %s", typeName, objID),
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Response is %+v", string(result)))

	// update the model with the new values
	var parsedCurrentDataPayload map[string]interface{}
	var parsedResponse map[string]interface{}

	err = json.Unmarshal(result, &parsedResponse)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Unmarshal object of type %s with id %s", typeName, objID),
			err.Error(),
		)
		return
	}

	err = json.Unmarshal([]byte(currentDataPayload), &parsedCurrentDataPayload)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Unmarshal object of type %s with id %s", typeName, objID),
			err.Error(),
		)
		return
	}

	// update only the fields which we provided
	// this is to avoid false updates due to the observability API generating new fields as response
	dataPayload := parsedResponse["data"].(map[string]interface{})
	for k, v := range dataPayload {
		if _, ok := parsedCurrentDataPayload[k]; ok {
			parsedCurrentDataPayload[k] = v
		}
	}

	// marshall the updated map into a json string to store in data attribute
	updatedDataPayload, err := json.Marshal(&parsedCurrentDataPayload)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Marshall object of type %s with id %s", typeName, objID),
			err.Error(),
		)
		return
	}

	// update the state data attribute
	data.Data = types.StringValue(string(updatedDataPayload))

	tflog.Debug(ctx, "read a resource")

	// set the placeholder value for testing purposses
	data.ID = types.StringValue("placeholder")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:gocritic // Terraform framework requires the method signature to be as is
func (r *ObjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Update method invoked")
	var data ObjectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// issue the API call
	typeName := data.TypeName.ValueString()
	objID := data.ObjectID.ValueString()
	layerType := data.LayerType.ValueString()
	layerID := data.LayerID.ValueString()
	jsonPayload := []byte(data.Data.ValueString())

	err := r.client.UpdateObject(typeName, objID, layerID, layerType, jsonPayload)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read type %s", typeName),
			err.Error(),
		)
		return
	}

	// set the placeholder value for testing purposses
	data.ID = types.StringValue("placeholder")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:gocritic // Terraform framework requires the method signature to be as is
func (r *ObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete method invoked")
	var data ObjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// issue the API call
	typeName := data.TypeName.ValueString()
	objID := data.ObjectID.ValueString()
	layerID := data.LayerID.ValueString()
	layerType := data.LayerType.ValueString()

	err := r.client.DeleteObject(typeName, objID, layerID, layerType)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Delete object of type %s with id %s", typeName, objID),
			err.Error(),
		)
		return
	}
}

func (r *ObjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
