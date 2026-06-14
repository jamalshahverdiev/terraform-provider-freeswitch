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

# Time-based routing: business hours -> queue; otherwise an "office closed"
# message. The first extension only matches Mon–Fri 09:00–17:59; outside that
# window its condition fails and the call falls through to the night one.
resource "freeswitch_dialplan_extension" "support_hours" {
  name     = "support-hours"
  domain   = freeswitch_domain.main.name
  context  = "company"
  priority = 20

  condition {
    field      = "destination_number"
    expression = "^(4444)$"
    time       = { wday = "2-6", hour = "9-17" } # wday 2-6 = Mon–Fri
    action {
      application = "transfer"
      data        = "support@${freeswitch_domain.main.name}"
    }
  }
}

resource "freeswitch_dialplan_extension" "support_closed" {
  name     = "support-closed"
  domain   = freeswitch_domain.main.name
  context  = "company"
  priority = 21

  condition {
    field      = "destination_number"
    expression = "^(4444)$"
    action {
      application = "answer"
    }
    action {
      application = "playback"
      data        = "ivr/office-closed.wav"
    }
  }
}
