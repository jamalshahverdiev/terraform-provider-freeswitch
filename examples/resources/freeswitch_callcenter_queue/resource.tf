resource "freeswitch_callcenter_queue" "support" {
  name = "support@192.168.48.143"

  # defaults: longest-idle-agent strategy, local_stream://moh,
  # discard-abandoned-after 60s
  strategy  = "longest-idle-agent"
  moh_sound = "local_stream://moh"
}
