resource "forwardemail_alias" "infracompany" {
  name   = "Infra Company"
  domain = "infracompany.com"

  recipients = ["james@rhodes.com"]

  description            = "Infra Company catch-all"
  has_imap               = true
  has_pgp                = false
  has_smime              = false
  has_wkd_disabled       = false
  error_code_if_disabled = 250
  max_quota              = "1 GB"
  retention              = 30
  smtp_limit             = 1000
}
