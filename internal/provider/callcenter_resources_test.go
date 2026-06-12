package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccCallcenterConfig(strategy string, noAnswerDelay int) string {
	return providerConfig + fmt.Sprintf(`
resource "freeswitch_callcenter_queue" "q" {
  name     = "tfacc-q@tfacc.example"
  strategy = %q

  params = {
    record-template = "/tmp/tfacc/${"$"}{uuid}.wav"
  }
}

resource "freeswitch_callcenter_agent" "a" {
  name                 = "tfacc-a@tfacc.example"
  contact              = "user/tfacc-a@tfacc.example"
  no_answer_delay_time = %d
}

resource "freeswitch_callcenter_tier" "t" {
  queue = freeswitch_callcenter_queue.q.name
  agent = freeswitch_callcenter_agent.a.name
  level = 2
}
`, strategy, noAnswerDelay)
}

func TestAccCallcenterResources(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create the full queue/agent/tier triple
				Config: testAccCallcenterConfig("longest-idle-agent", 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeswitch_callcenter_queue.q", "strategy", "longest-idle-agent"),
					// defaults come back computed
					resource.TestCheckResourceAttr("freeswitch_callcenter_queue.q", "moh_sound", "local_stream://moh"),
					resource.TestCheckResourceAttr("freeswitch_callcenter_queue.q", "discard_abandoned_after", "60"),
					resource.TestCheckResourceAttr("freeswitch_callcenter_agent.a", "type", "callback"),
					resource.TestCheckResourceAttr("freeswitch_callcenter_agent.a", "status", "Available"),
					resource.TestCheckResourceAttr("freeswitch_callcenter_agent.a", "no_answer_delay_time", "5"),
					resource.TestCheckResourceAttr("freeswitch_callcenter_tier.t", "level", "2"),
					resource.TestCheckResourceAttr("freeswitch_callcenter_tier.t", "position", "1"),
				),
			},
			{ // import queue by name
				ResourceName:      "freeswitch_callcenter_queue.q",
				ImportState:       true,
				ImportStateId:     "tfacc-q@tfacc.example",
				ImportStateVerify: true,
			},
			{ // import tier by "queue/agent"
				ResourceName:      "freeswitch_callcenter_tier.t",
				ImportState:       true,
				ImportStateId:     "tfacc-q@tfacc.example/tfacc-a@tfacc.example",
				ImportStateVerify: true,
			},
			{ // in-place update of queue strategy + agent timer
				Config: testAccCallcenterConfig("round-robin", 7),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeswitch_callcenter_queue.q", "strategy", "round-robin"),
					resource.TestCheckResourceAttr("freeswitch_callcenter_agent.a", "no_answer_delay_time", "7"),
				),
			},
		},
	})
}
