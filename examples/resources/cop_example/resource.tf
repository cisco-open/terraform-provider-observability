terraform {
  required_providers {
    cop = {
      source = "testTerraform.com/appd/cop"
    }
  }
}

provider "cop" {
  tenant      = "47a01df9-54a0-472b-96b8-7c8f64eb7cbf"
  auth_method = "oauth"
  url         = "https://alameda-c0-test-02.saas.appd-test.com"
}

resource "cop_object" "ns" {
  type_name = "fmm:namespace"
  object_id = "aws"
  layer_type = "TENANT"
  layer_id = "47a01df9-54a0-472b-96b8-7c8f64eb7cbf"
}
