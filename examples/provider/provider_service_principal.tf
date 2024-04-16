# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.
#
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    observability = {
      source = "testTerraform.com/appd/observability"
    }
  }
}

provider "observability" {
  tenant="47a01df9-54a0-472b-96b8-7c8f64eb7cbf"
  auth_method="service-principal"
  url="https://alameda-c0-test-02.saas.appd-test.com"
  secrets_file="/home/vdodin/secrets.json"
}

