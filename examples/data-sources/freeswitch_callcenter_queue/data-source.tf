# Reference a queue owned by another Terraform state / created via API.
data "freeswitch_callcenter_queue" "support" {
  name = "support@192.168.48.143"
}

# e.g. bind an extra agent to it
resource "freeswitch_callcenter_tier" "extra" {
  queue = data.freeswitch_callcenter_queue.support.name
  agent = freeswitch_callcenter_agent.backup.name
}
