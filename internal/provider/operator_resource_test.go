package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccOperatorConfig(displayName string, enabled bool) string {
	return providerConfig + fmt.Sprintf(`
resource "freeswitch_operator" "o" {
  subject      = "kc-acc-subject-1"
  domain       = "tfacc.example"
  number       = "6010"
  display_name = %q
  enabled      = %t
}
`, displayName, enabled)
}

func TestAccOperatorResource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create
				Config: testAccOperatorConfig("Agent One", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeswitch_operator.o", "subject", "kc-acc-subject-1"),
					resource.TestCheckResourceAttr("freeswitch_operator.o", "domain", "tfacc.example"),
					resource.TestCheckResourceAttr("freeswitch_operator.o", "number", "6010"),
					resource.TestCheckResourceAttr("freeswitch_operator.o", "display_name", "Agent One"),
					resource.TestCheckResourceAttr("freeswitch_operator.o", "enabled", "true"),
					resource.TestCheckResourceAttrSet("freeswitch_operator.o", "id"),
				),
			},
			{ // import by subject
				ResourceName:      "freeswitch_operator.o",
				ImportState:       true,
				ImportStateId:     "kc-acc-subject-1",
				ImportStateVerify: true,
			},
			{ // update in place
				Config: testAccOperatorConfig("Reception", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeswitch_operator.o", "display_name", "Reception"),
					resource.TestCheckResourceAttr("freeswitch_operator.o", "enabled", "false"),
				),
			},
			{ // data source reads it back
				Config: testAccOperatorConfig("Reception", false) + `
data "freeswitch_operator" "by_subject" {
  subject = freeswitch_operator.o.subject
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeswitch_operator.by_subject", "number", "6010"),
					resource.TestCheckResourceAttr("data.freeswitch_operator.by_subject", "display_name", "Reception"),
				),
			},
		},
	})
}
