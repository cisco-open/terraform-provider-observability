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
  tenant="0eb4e853-34fb-4f77-b3fc-b9cd3b462366"
  auth_method="service-principal"
  url="https://aiops-dev.saas.appd-test.com"
  secrets_file="/home/vdodin/aiops_secret.json"
}


resource "observability_object" "conn" {
  type_name = "anzen:cloudConnection"
  object_id = "just-terraform-testing"
  layer_type = "TENANT"
  layer_id = "0eb4e853-34fb-4f77-b3fc-b9cd3b462366"
  data = jsonencode(
    {
      "cloudType": "AWS",
      "connectionName": "just-terraform-testing",
      "region": "us-east-2",
      "accessKey": "**********",
      "secretAccessKey": "**********",
      "s3AccessLogBucket": "s3://s3-sanity-logging/",
      "athenaOutputBucket": "s3://s3-sanity-athena-logs/",
      "createTimestamp": "",
      "accountId": "81892134343434"
    }
  )
}
