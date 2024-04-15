// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type IsValidJSONString struct {
	JSONString string
}

func (v IsValidJSONString) Description(_ context.Context) string {
	return fmt.Sprintf("value %s must be a valid json encoding string", v.JSONString)
}

func (v IsValidJSONString) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

//nolint:gocritic // Terraform framework requires the method signature to be as is
func (v IsValidJSONString) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue

	if json.Valid([]byte(value.ValueString())) {
		return
	}

	v.JSONString = value.ValueString()

	response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
		request.Path,
		v.Description(ctx),
		value.String(),
	))
}
