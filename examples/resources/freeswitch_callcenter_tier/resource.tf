resource "freeswitch_callcenter_tier" "support_4201" {
  queue = freeswitch_callcenter_queue.support.name
  agent = freeswitch_callcenter_agent.agent_4201.name
  level = 1
}
