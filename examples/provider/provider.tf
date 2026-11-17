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

  # Domain deletion is destructive and irreversible; it is refused unless
  # explicitly allowed here (or FORWARDEMAIL_ALLOW_DOMAIN_DELETION=true).
  # allow_domain_deletion = true
}
