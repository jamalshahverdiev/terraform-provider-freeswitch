package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccDomainConfig(name, desc string) string {
	return providerConfig + fmt.Sprintf(`
resource "freeswitch_domain" "test" {
  name        = %q
  description = %q
  variables   = { default_language = "en" }
}
`, name, desc)
}

func TestAccDomainResource(t *testing.T) {
	testAccPreCheck(t)
	const name = "tfacc.example"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create
				Config: testAccDomainConfig(name, "first"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeswitch_domain.test", "name", name),
					resource.TestCheckResourceAttr("freeswitch_domain.test", "description", "first"),
					resource.TestCheckResourceAttr("freeswitch_domain.test", "enabled", "true"),
					resource.TestCheckResourceAttr("freeswitch_domain.test", "variables.default_language", "en"),
					resource.TestCheckResourceAttrSet("freeswitch_domain.test", "id"),
					resource.TestCheckResourceAttrSet("freeswitch_domain.test", "created_at"),
				),
			},
			{ // import
				ResourceName:      "freeswitch_domain.test",
				ImportState:       true,
				ImportStateId:     name,
				ImportStateVerify: true,
			},
			{ // update (in place)
				Config: testAccDomainConfig(name, "second"),
				Check: resource.TestCheckResourceAttr("freeswitch_domain.test", "description", "second"),
			},
		},
	})
}
