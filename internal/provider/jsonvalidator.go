package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type IsValidJsonString struct {
	JsonString string
}

func (v IsValidJsonString) Description(_ context.Context) string {
	return fmt.Sprintf("value %s must be a valid json encoding string", v.JsonString)
}

func (v IsValidJsonString) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v IsValidJsonString) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue

	if json.Valid([]byte(value.ValueString())) {
		return
	}

	v.JsonString = value.ValueString()

	response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
		request.Path,
		v.Description(ctx),
		value.String(),
	))
}
