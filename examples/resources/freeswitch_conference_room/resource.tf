# A room also materializes the dialplan extension callers dial to enter it.
resource "freeswitch_conference_room" "standup" {
  name    = "standup"
  number  = "3500"
  domain  = "192.168.48.143"
  context = "company"
  profile = freeswitch_conference_profile.video_grid.name
}

# PIN-protected room.
resource "freeswitch_conference_room" "private" {
  name    = "private"
  number  = "3501"
  domain  = "192.168.48.143"
  context = "company"
  profile = freeswitch_conference_profile.video_grid.name
  pin     = "1234"
}
