package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Covers time-based routing: a condition with a regex + time window, and a
// pure time gate (no field/expression). Requires a domain to attach to.
func testAccTimeRoutingConfig() string {
	return providerConfig + `
resource "freeswitch_domain" "d" { name = "tfacc-tr.example" }

resource "freeswitch_dialplan_extension" "hours" {
  name     = "tfacc-hours"
  domain   = freeswitch_domain.d.name
  context  = "tfacc-tr"
  priority = 10
  condition {
    field      = "destination_number"
    expression = "^(4444)$"
    time       = { wday = "2-6", hour = "9-17" }
    action {
      application = "transfer"
      data        = "support@tfacc-tr.example"
    }
  }
}

resource "freeswitch_dialplan_extension" "night" {
  name     = "tfacc-night"
  domain   = freeswitch_domain.d.name
  context  = "tfacc-tr"
  priority = 11
  condition {
    time = { "time-of-day" = "17:00-9:00" }
    action {
      application = "answer"
    }
    action {
      application = "playback"
      data        = "closed.wav"
    }
  }
}
`
}

func TestAccDialplanTimeRouting(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTimeRoutingConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeswitch_dialplan_extension.hours", "condition.0.time.wday", "2-6"),
					resource.TestCheckResourceAttr("freeswitch_dialplan_extension.hours", "condition.0.time.hour", "9-17"),
					resource.TestCheckResourceAttr("freeswitch_dialplan_extension.hours", "condition.0.field", "destination_number"),
					// pure time gate: no field, time present
					resource.TestCheckNoResourceAttr("freeswitch_dialplan_extension.night", "condition.0.field"),
					resource.TestCheckResourceAttr("freeswitch_dialplan_extension.night", "condition.0.time.time-of-day", "17:00-9:00"),
				),
			},
		},
	})
}
