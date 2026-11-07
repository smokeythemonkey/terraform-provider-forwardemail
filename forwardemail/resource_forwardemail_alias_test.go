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
					resource.TestCheckResourceAttr("forwardemail_domain.test", "name", "example.com"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "domain", "example.com"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "name", "postmaster"),
				),
			},
			{
				Config: testAccAliasResourceConfigUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("forwardemail_alias.test", "recipients.#", "2"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "recipients.0", "user@example.com"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "recipients.1", "admin@example.com"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "description", "updated description"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "labels.#", "1"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "labels.0", "production"),
					resource.TestCheckResourceAttr("forwardemail_alias.test", "enabled", "false"),
				),
			},
			{
				ResourceName:      "forwardemail_alias.test",
				ImportState:       true,
				ImportStateId:     "example.com/postmaster",
				ImportStateVerify: true,
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
  name = "example.com"
}

resource "forwardemail_alias" "test" {
  name       = "postmaster"
  domain     = forwardemail_domain.test.name
  recipients = ["user@example.com"]
}
`
}

func testAccAliasResourceConfigUpdate() string {
	return `
resource "forwardemail_domain" "test" {
  name = "example.com"
}

resource "forwardemail_alias" "test" {
  name        = "postmaster"
  domain      = forwardemail_domain.test.name
  recipients  = ["user@example.com", "admin@example.com"]
  description = "updated description"
  labels      = ["production"]
  enabled     = false
}
`
}
