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


data "cop_type" "ns" {
  type_name = "anzen:cloudConnection"
}

output "myType" {
  value = data.cop_type.ns
}
