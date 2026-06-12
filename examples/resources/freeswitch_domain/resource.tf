resource "freeswitch_domain" "main" {
  name        = "192.168.48.143"
  description = "Main PBX domain"

  variables = {
    default_language = "en"
  }
}
