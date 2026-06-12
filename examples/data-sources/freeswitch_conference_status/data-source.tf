# Live participants of a running conference (running=false when idle).
data "freeswitch_conference_status" "standup" {
  name = "standup"
}

output "standup_members" {
  value = [for m in data.freeswitch_conference_status.standup.members : m.caller_id_number]
}
