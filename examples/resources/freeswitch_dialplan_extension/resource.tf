resource "freeswitch_dialplan_extension" "internal" {
  name     = "internal-2xxx"
  domain   = freeswitch_domain.main.name
  context  = "company"
  priority = 10

  condition {
    field      = "destination_number"
    expression = "^(20[0-9][0-9])$"

    action {
      application = "bridge"
      data        = "user/$1@${freeswitch_domain.main.name}"
    }
  }
}
