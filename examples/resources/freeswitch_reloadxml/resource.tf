# Flushes the FreeSWITCH XML cache whenever the triggers change.
resource "freeswitch_reloadxml" "apply" {
  triggers = {
    ext = freeswitch_dialplan_extension.ivr.id
  }
}
