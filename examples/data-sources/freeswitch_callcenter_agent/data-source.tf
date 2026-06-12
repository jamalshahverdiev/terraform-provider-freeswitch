data "freeswitch_callcenter_agent" "agent_4201" {
  name = "4201@192.168.48.143"
}

output "agent_contact" {
  value = data.freeswitch_callcenter_agent.agent_4201.contact
}
