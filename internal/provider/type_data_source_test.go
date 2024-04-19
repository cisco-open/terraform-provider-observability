// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//lint:ignore U1000 Ignore unused function temporarily for debugging
func _TestAccTypeDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "observability_type" "test" {
					type_name = "fmm:namespace"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.observability_type.test", "type_name", "fmm:namespace"),

					//nolint:lll // Due to payload nature of data field
					resource.TestCheckResourceAttr("data.observability_type.test", "data", "{\"jsonSchema\":{\"type\":\"object\",\"title\":\"Namespace object\",\"$schema\":\"http://json-schema.org/draft-07/schema\",\"default\":{},\"examples\":[{\"name\":\"apm\"}],\"required\":[\"name\"],\"properties\":{\"name\":{\"type\":\"string\",\"title\":\"Namespace name\",\"pattern\":\"^[a-z][a-z0-9_.]{0,36}$\",\"maxLength\":36,\"minLength\":1}},\"additionalProperties\":false},\"idGeneration\":{\"generateRandomId\":false,\"idGenerationMechanism\":\"{{object.name}}\",\"enforceGlobalUniqueness\":true},\"allowObjectFragments\":false,\"allowedLayers\":[\"SOLUTION\"],\"solution\":\"fmm\",\"name\":\"namespace\",\"createdAt\":\"2023-03-29T20:44:14.495Z\",\"updatedAt\":\"2024-01-31T01:23:15.492Z\"}"),

					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.observability_type.test", "id", "placeholder"),
				),
			},
		},
	})
}
