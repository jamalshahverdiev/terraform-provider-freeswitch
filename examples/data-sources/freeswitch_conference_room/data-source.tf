data "freeswitch_conference_room" "standup" {
  name = "standup"
}

output "standup_number" {
  value = data.freeswitch_conference_room.standup.number
}
