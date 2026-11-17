package forwardemail

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccDomainResource(t *testing.T) resource.TestCase {
	t.Setenv("FORWARDEMAIL_ALLOW_DOMAIN_DELETION", "true")
	return resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("forwardemail_domain.test", "name", "forwardemailprovider.com"),
				),
			},
			{
				Config: testAccDomainResourceConfigUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("forwardemail_domain.test", "name", "forwardemailprovider.com"),
					resource.TestCheckResourceAttr("forwardemail_domain.test", "adult_content_protection", "false"),
					resource.TestCheckResourceAttr("forwardemail_domain.test", "recipient_verification", "false"),
				),
			},
			{
				ResourceName:      "forwardemail_domain.test",
				ImportState:       true,
				ImportStateId:     "forwardemailprovider.com",
				ImportStateVerify: true,
			},
		},
	}
}

func testAccDomainResourceConfig() string {
	return `
resource "forwardemail_domain" "test" {
  name = "forwardemailprovider.com"
}
`
}

func testAccDomainResourceConfigUpdate() string {
	return `
resource "forwardemail_domain" "test" {
  name = "forwardemailprovider.com"

  adult_content_protection = false
  recipient_verification    = false
}
`
}
