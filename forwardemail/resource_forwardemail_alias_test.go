package forwardemail

import (
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
