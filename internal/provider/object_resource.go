// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	ImportID  types.String `tfsdk:"import_id"`
	ID        types.String `tfsdk:"id"`
}

func (r *ObjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object"
}

// Schema defines the schema for the resource.
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
					IsValidJSONString{},
				},
			},
			"import_id": schema.StringAttribute{
				MarkdownDescription: "ID used when doing import operation on an object",
				Optional:            true,
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

//nolint:gocritic,funlen // Terraform framework requires the method signature to be as is
func (r *ObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read method invoked")
	var data ObjectResourceModel
	var importIDTokenLength = 4

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
	importIdentifier := data.ImportID.ValueString()

	// in case of import previous properties will be empty, only importIdentifier will get populated
	// because of this, importIdentifier will be a composition of typeName|objID|layerID|layerType
	// reason for this is that the api needs all four fields to properly identify an object
	identityFields := strings.Split(importIdentifier, "|")
	if len(identityFields) == importIDTokenLength {
		tflog.Debug(ctx, "Import scenario detected")
		tflog.Debug(ctx, "Extracting required fields from import_id")
		typeName = identityFields[0]
		objID = identityFields[1]
		layerType = identityFields[2]
		layerID = identityFields[3]
	}

	tflog.Debug(ctx, fmt.Sprintf("type name is %s", typeName))
	tflog.Debug(ctx, fmt.Sprintf("object id is %s", objID))
	tflog.Debug(ctx, fmt.Sprintf("layer ID is %s", layerID))
	tflog.Debug(ctx, fmt.Sprintf("layer type %s", layerType))
	tflog.Debug(ctx, fmt.Sprintf("data payload %s", currentDataPayload))
	tflog.Debug(ctx, fmt.Sprintf("import identifier is %s", importIdentifier))

	result, err := r.client.GetObject(typeName, objID, layerID, layerType)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read object of type %s with id %s", typeName, objID),
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("api response is %s", string(result)))
	// update the model with the new values
	var parsedCurrentDataPayload map[string]any
	var parsedResponse map[string]any

	err = json.Unmarshal(result, &parsedResponse)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Unmarshal response object of type %s with id %s", typeName, objID),
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("parsed response into map is %v", parsedResponse))
	if currentDataPayload != "" {
		err = json.Unmarshal([]byte(currentDataPayload), &parsedCurrentDataPayload)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to Unmarshal current object of type %s with id %s", typeName, objID),
				err.Error(),
			)
			return
		}
	}

	// if we can't fetch any data from the cloud return
	var dataPayload map[string]any
	var ok bool
	if dataPayload, ok = parsedResponse["data"].(map[string]any); !ok {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to assert data map from current object of type %s with id %s", typeName, objID),
			err.Error(),
		)
		return
	}

	if parsedCurrentDataPayload == nil {
		// data was not provided in this case, maybe import usecase
		// populate all the fields with what the observability api provided
		parsedCurrentDataPayload = dataPayload
	} else {
		// update only the fields which we provided
		// this is to avoid false updates due to the observability API generating new fields as response
		for k, v := range dataPayload {
			if _, ok := parsedCurrentDataPayload[k]; ok {
				parsedCurrentDataPayload[k] = v
			}
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
	// set the rest of the fields
	data.TypeName = types.StringValue(typeName)
	data.ObjectID = types.StringValue(objID)
	data.LayerType = types.StringValue(layerType)
	data.LayerID = types.StringValue(layerID)

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
	resource.ImportStatePassthroughID(ctx, path.Root("import_id"), req, resp)
}
