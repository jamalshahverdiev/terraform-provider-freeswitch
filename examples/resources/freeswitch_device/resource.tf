# A provisioned desk phone. The SIP password is NOT stored here — the
# control-plane reads it from the matching freeswitch_user at render time and
# serves the vendor config at GET /provision/<mac> (behind Basic auth + CIDR).

resource "freeswitch_user" "reception" {
  domain   = "example.com"
  number   = "1001"
  password = var.reception_sip_password
}

resource "freeswitch_device" "reception" {
  mac          = "80:5e:c0:11:22:33" # any separators; normalized server-side
  vendor       = "yealink"           # yealink | grandstream | generic
  model        = "T46U"
  number       = freeswitch_user.reception.number
  domain       = freeswitch_user.reception.domain
  display_name = "Reception"
  enabled      = true
}

# A Grandstream phone for the same extension family. The control-plane renders
# its config as Grandstream P-value XML when the phone fetches cfg<mac>.xml.
resource "freeswitch_user" "lobby" {
  domain   = "example.com"
  number   = "1002"
  password = var.lobby_sip_password
}

resource "freeswitch_device" "lobby" {
  mac          = "00:0b:82:aa:bb:cc"
  vendor       = "grandstream"
  model        = "GXP2170"
  number       = freeswitch_user.lobby.number
  domain       = freeswitch_user.lobby.domain
  display_name = "Lobby"
}
