package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccDeviceConfig(vendor, displayName string, enabled bool) string {
	return providerConfig + fmt.Sprintf(`
resource "freeswitch_device" "d" {
  mac          = "001122334455"
  vendor       = %q
  number       = "1099"
  domain       = "tfacc.example"
  display_name = %q
  enabled      = %t
}
`, vendor, displayName, enabled)
}

func TestAccDeviceResource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create — MAC normalized to lowercase, no separators
				Config: testAccDeviceConfig("yealink", "Reception", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeswitch_device.d", "mac", "001122334455"),
					resource.TestCheckResourceAttr("freeswitch_device.d", "model", ""),
					resource.TestCheckResourceAttr("freeswitch_device.d", "vendor", "yealink"),
					resource.TestCheckResourceAttr("freeswitch_device.d", "number", "1099"),
					resource.TestCheckResourceAttr("freeswitch_device.d", "display_name", "Reception"),
					resource.TestCheckResourceAttr("freeswitch_device.d", "enabled", "true"),
					resource.TestCheckResourceAttrSet("freeswitch_device.d", "id"),
				),
			},
			{ // import by MAC
				ResourceName:      "freeswitch_device.d",
				ImportState:       true,
				ImportStateId:     "001122334455",
				ImportStateVerify: true,
			},
			{ // update in place: vendor, name, enabled
				Config: testAccDeviceConfig("grandstream", "Lobby", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeswitch_device.d", "vendor", "grandstream"),
					resource.TestCheckResourceAttr("freeswitch_device.d", "display_name", "Lobby"),
					resource.TestCheckResourceAttr("freeswitch_device.d", "enabled", "false"),
				),
			},
			{ // data source reads the same device back by mac
				Config: testAccDeviceConfig("grandstream", "Lobby", false) + `
data "freeswitch_device" "by_mac" {
  mac = freeswitch_device.d.mac
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeswitch_device.by_mac", "number", "1099"),
					resource.TestCheckResourceAttr("data.freeswitch_device.by_mac", "vendor", "grandstream"),
					resource.TestCheckResourceAttr("data.freeswitch_device.by_mac", "display_name", "Lobby"),
					resource.TestCheckResourceAttr("data.freeswitch_device.by_mac", "enabled", "false"),
				),
			},
		},
	})
}
