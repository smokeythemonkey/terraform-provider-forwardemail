package forwardemail

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccAliasResource(t *testing.T) resource.TestCase {
	return resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasResourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("forwardemail_domain.test", "name", "forwardemailprovider.com"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "domain", "forwardemailprovider.com"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "name", "postmaster"),
				),
			},
			{
				Config: testAccAliasResourceConfigUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("forwardemail_alias.test", "recipients.#", "2"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "recipients.0", "user@external-example.com"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "recipients.1", "admin@external-example.com"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "description", "updated description"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "labels.#", "1"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "labels.0", "production"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "enabled", "false"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "has_imap", "true"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "has_pgp", "true"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "has_smime", "false"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "has_wkd_disabled", "true"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "error_code_if_disabled", "421"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "max_quota", "1 GB"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "retention", "30"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "smtp_limit", "100"),
				),
			},
			{
				ResourceName:      "forwardemail_alias.test",
				ImportState:       true,
				ImportStateId:     "forwardemailprovider.com/postmaster",
				ImportStateVerify: true,
				// One-way params are accepted by the API but not echoed in
				// GET responses, so imported state cannot reproduce them.
				ImportStateVerifyIgnore: []string{
					"error_code_if_disabled",
					"max_quota",
					"public_key",
					"smime_certificate",
				},
			},
			{
				ResourceName:  "forwardemail_alias.test",
				ImportState:   true,
				ImportStateId: "oops",
				ExpectError:   regexp.MustCompile(`Invalid import ID`),
			},
		},
	}
}

func testAccAliasResourceConfig() string {
	return `
resource "forwardemail_domain" "test" {
  name = "forwardemailprovider.com"
}

resource "forwardemail_alias" "test" {
  name       = "postmaster"
  domain     = forwardemail_domain.test.name
  recipients = ["user@external-example.com"]
}
`
}

func testAccAliasResourceConfigUpdate() string {
	return `
resource "forwardemail_domain" "test" {
  name = "forwardemailprovider.com"
}

resource "forwardemail_alias" "test" {
  name        = "postmaster"
  domain      = forwardemail_domain.test.name
  recipients  = ["user@external-example.com", "admin@external-example.com"]
  description = "updated description"
  labels      = ["production"]
  enabled     = false

  has_imap               = true
  has_pgp                = true
  has_smime              = false
  has_wkd_disabled       = true
  error_code_if_disabled = 421
  max_quota              = "1 GB"
  retention              = 30
  smtp_limit             = 100
}
`
}
