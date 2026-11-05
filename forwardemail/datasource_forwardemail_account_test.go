package forwardemail

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccDataSourceForwardemailAccount(t *testing.T) resource.TestCase {
	return resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.forwardemail_account.test", "email"),
				),
			},
		},
	}
}

func testAccAccountDataSourceConfig() string {
	return `
data forwardemail_account "test" {}`
}
