terraform {
  required_version = ">= 1.0.0"

  required_providers {
    forwardemail = {
      source  = "smokeythemonkey/forwardemail"
      version = ">= 0.1"
    }
  }
}

provider "forwardemail" {
  api_key = "XXXXXXX"
}
