// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package provider_test

import (
	"context"
	"testing"

	"github.com/cisco-open/terraform-provider-observability/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// unit test for IsValidJsonString
func TestIsValidJSONString(t *testing.T) {
	myValidator := provider.IsValidJSONString{}

	tests := []struct {
		input         string
		expectedValid bool
	}{
		{"{\"key\": \"value\"}", true},               // valid JSON
		{"{\"key\": \"value\"", false},               // invalid JSON (missing closing bracket)
		{"invalid JSON string", false},               // invalid JSON
		{"{\"key\": {\"nested\": true}}", true},      // valid JSON (nested)
		{"{\"key\": [1,2,3], \"nested\": {}}", true}, // valid JSON (nested with array)
		//nolint:lll // Due to payload nature of data field
		{"{\"layerType\":\"TENANT\",\"id\":\"just-terraform-testing\",\"layerId\":\"0eb4e853-34fb-4f77-b3fc-b9cd3b462366\",\"data\":{\"region\":\"us-east-2\",\"accessKey\":\"myAccKey\",\"accountId\":\"81892134343434\",\"cloudType\":\"AWS\",\"connectionName\":\"just-terraform-testing\",\"createTimestamp\":\"\",\"secretAccessKey\":\"mySecretAccKey\",\"s3AccessLogBucket\":\"s3://s3-sanity-logging/\",\"athenaOutputBucket\":\"s3://s3-sanity-athena-logs/\"},\"objectMimeType\":\"application/json\",\"targetObjectId\":null,\"patch\":null,\"createdAt\":\"2024-04-29T11:54:28.456Z\",\"updatedAt\":\"2024-04-29T11:54:28.456Z\",\"objectType\":\"anzen:cloudConnection\"}", true}, // valid JSON example of actual terraform configuration
		//nolint:lll // Due to payload nature of data field
		{"{\"layerType\":\"TENANT\",\"id\":\"just-terraform-testing\",\"layerId\":\"0eb4e853-34fb-4f77-b3fc-b9cd3b462366\",\"data\":{\"region\":\"us-east-2\",\"accessKey\":\"myAccKey\",\"accountId\":\"81892134343434\",\"cloudType\":\"AWS\",\"connectionName\":\"just-terraform-testing\",\"createTimestamp\":\"\",\"secretAccessKey\":\"mySecretAccKey\",\"s3AccessLogBucket\":\"s3://s3-sanity-logging/\",\"athenaOutputBucket\":\"s3://s3-sanity-athena-logs/\"},\"objectMimeType\":\"application/json\",,,,\"targetObjectId\":null,\"patch\":null,\"createdAt\":\"2024-04-29T11:54:28.456Z\",\"updatedAt\":\"2024-04-29T11:54:28.456Z\",\"objectType\":\"anzen:cloudConnection\"}", false}, // invalid JSON example of actual terraform configuration

		{"{\"key\": 123}", true},           // valid JSON (number value)
		{"[\"value1\", \"value2\"]", true}, // valid JSON (array)
		{"true", true},                     // valid JSON (boolean)
		{"null", true},                     // valid JSON (null)
	}

	for _, test := range tests {
		req := validator.StringRequest{
			ConfigValue: types.StringValue(test.input),
		}
		resp := &validator.StringResponse{}

		myValidator.ValidateString(context.Background(), req, resp)

		if test.expectedValid && resp.Diagnostics.HasError() {
			t.Errorf("Expected '%s' to be valid, but got errors: %v", test.input, resp.Diagnostics)
		} else if !test.expectedValid && !resp.Diagnostics.HasError() {
			t.Errorf("Expected '%s' to be invalid, but got no errors", test.input)
		}
	}
}
