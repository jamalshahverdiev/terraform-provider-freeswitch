# Reloads mod_callcenter (re-reads queues/agents/tiers from the control
# plane) whenever the triggers change. Plain reloadxml is NOT enough.
resource "freeswitch_callcenter_reload" "apply" {
  triggers = {
    queue  = freeswitch_callcenter_queue.support.id
    agents = join(",", [for a in freeswitch_callcenter_agent.agent : a.id])
    tiers  = join(",", [for t in freeswitch_callcenter_tier.tier : t.id])
  }
}
