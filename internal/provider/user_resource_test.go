package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccUserVoicemailConfig(email string, attach bool) string {
	return providerConfig + fmt.Sprintf(`
resource "freeswitch_domain" "d" { name = "tfacc-vm.example" }

resource "freeswitch_user" "u" {
  domain = freeswitch_domain.d.name
  number = "6010"
  params = { password = "sip-secret" }
  voicemail = {
    password    = "4321"
    email       = %q
    attach_file = %t
  }
}
`, email, attach)
}

func TestAccVoicemailDataSource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // a user with no messages yields an empty mailbox (counters = 0)
				Config: providerConfig + `
data "freeswitch_voicemail" "vm" {
  domain = "tfacc-vm.example"
  number = "9999"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeswitch_voicemail.vm", "total", "0"),
					resource.TestCheckResourceAttr("data.freeswitch_voicemail.vm", "unread", "0"),
					resource.TestCheckResourceAttr("data.freeswitch_voicemail.vm", "messages.#", "0"),
				),
			},
		},
	})
}

func TestAccUserVoicemail(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create with a typed voicemail block; enabled/email_all default
				Config: testAccUserVoicemailConfig("vm@example.com", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeswitch_user.u", "number", "6010"),
					resource.TestCheckResourceAttr("freeswitch_user.u", "voicemail.enabled", "true"),
					resource.TestCheckResourceAttr("freeswitch_user.u", "voicemail.password", "4321"),
					resource.TestCheckResourceAttr("freeswitch_user.u", "voicemail.email", "vm@example.com"),
					resource.TestCheckResourceAttr("freeswitch_user.u", "voicemail.attach_file", "true"),
					resource.TestCheckResourceAttr("freeswitch_user.u", "voicemail.email_all", "false"),
				),
			},
			{ // import by domain/number — voicemail round-trips from the API
				ResourceName:      "freeswitch_user.u",
				ImportState:       true,
				ImportStateId:     "tfacc-vm.example/6010",
				ImportStateVerify: true,
				// params is not echoed by the API, so it can't be import-verified
				ImportStateVerifyIgnore: []string{"params"},
			},
			{ // update the mailbox in place
				Config: testAccUserVoicemailConfig("changed@example.com", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeswitch_user.u", "voicemail.email", "changed@example.com"),
					resource.TestCheckResourceAttr("freeswitch_user.u", "voicemail.attach_file", "false"),
				),
			},
			{ // data source reads the mailbox back
				Config: testAccUserVoicemailConfig("changed@example.com", false) + `
data "freeswitch_user" "by_id" {
  domain = freeswitch_user.u.domain
  number = freeswitch_user.u.number
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeswitch_user.by_id", "voicemail.password", "4321"),
					resource.TestCheckResourceAttr("data.freeswitch_user.by_id", "voicemail.email", "changed@example.com"),
				),
			},
		},
	})
}
