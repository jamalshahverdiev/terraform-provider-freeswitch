resource "freeswitch_gateway" "provider_main" {
  name     = "provider-main"
  profile  = "external"
  proxy    = "sip.provider.example"
  realm    = "sip.provider.example"
  username = "sip-user"
  password = var.sip_password
  register = true

  params = {
    expire-seconds = "3600"
  }
}
