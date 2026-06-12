data "freeswitch_callcenter_tier" "support_4201" {
  queue = "support@192.168.48.143"
  agent = "4201@192.168.48.143"
}

output "tier_level" {
  value = data.freeswitch_callcenter_tier.support_4201.level
}
