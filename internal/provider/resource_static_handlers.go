// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package provider

import "github.com/hashicorp/terraform-plugin-framework/resource"

func registerStaticResourceHandlers() []func() resource.Resource {
	// for resources that are not generated register them here
	return []func() resource.Resource{}
}
