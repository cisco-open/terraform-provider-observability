// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

//go:build acceptance

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `
terraform {
	required_providers {
		observability = {
		source = "registry.terraform.io/cisco-open/observability",
		}
	}
}

provider "observability" {
	tenant="0eb4e853-34fb-4f77-b3fc-b9cd3b462366"
	auth_method="service-principal"
	url="https://aiops-dev.saas.appd-test.com"
	secrets_file="/home/vdodin/aiops_secret.json"
}
`
)

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"observability": providerserver.NewProtocol6WithError(New("test")()),
	}
)
