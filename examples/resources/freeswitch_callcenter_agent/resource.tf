resource "freeswitch_callcenter_agent" "agent_4201" {
  name    = "4201@192.168.48.143"
  contact = "user/4201@192.168.48.143"
  status  = "Available"

  # for demos: retry an unanswered agent quickly
  no_answer_delay_time = 5
}
