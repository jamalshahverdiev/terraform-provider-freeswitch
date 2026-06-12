package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccConferenceConfig(videoMode, pin string) string {
	return providerConfig + fmt.Sprintf(`
resource "freeswitch_conference_profile" "p" {
  name       = "tfacc-profile"
  video_mode = %q
}

resource "freeswitch_conference_room" "r" {
  name    = "tfacc-room"
  number  = "9991"
  domain  = "tfacc.example"
  context = "tfacc-ctx"
  profile = freeswitch_conference_profile.p.name
  pin     = %q
}
`, videoMode, pin)
}

func TestAccConferenceResources(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create video profile + pin room
				Config: testAccConferenceConfig("mux", "4321"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeswitch_conference_profile.p", "video_mode", "mux"),
					// computed defaults
					resource.TestCheckResourceAttr("freeswitch_conference_profile.p", "video_layout", "group:grid"),
					resource.TestCheckResourceAttr("freeswitch_conference_profile.p", "video_canvas_size", "1280x720"),
					resource.TestCheckResourceAttr("freeswitch_conference_profile.p", "rate", "48000"),
					resource.TestCheckResourceAttr("freeswitch_conference_room.r", "number", "9991"),
					resource.TestCheckResourceAttr("freeswitch_conference_room.r", "pin", "4321"),
					resource.TestCheckResourceAttr("freeswitch_conference_room.r", "priority", "5"),
					resource.TestCheckResourceAttr("freeswitch_conference_room.r", "enabled", "true"),
				),
			},
			{ // import room by name
				ResourceName:      "freeswitch_conference_room.r",
				ImportState:       true,
				ImportStateId:     "tfacc-room",
				ImportStateVerify: true,
			},
			{ // import profile by name
				ResourceName:      "freeswitch_conference_profile.p",
				ImportState:       true,
				ImportStateId:     "tfacc-profile",
				ImportStateVerify: true,
			},
			{ // updates in place: drop video, change pin
				Config: testAccConferenceConfig("", "9876"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeswitch_conference_profile.p", "video_mode", ""),
					resource.TestCheckResourceAttr("freeswitch_conference_room.r", "pin", "9876"),
				),
			},
		},
	})
}
