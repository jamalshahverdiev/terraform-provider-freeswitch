# Reuse a shared profile owned by another Terraform state.
data "freeswitch_conference_profile" "video_grid" {
  name = "video-grid"
}

resource "freeswitch_conference_room" "team_b" {
  name    = "team-b"
  number  = "3502"
  domain  = "192.168.48.143"
  context = "company"
  profile = data.freeswitch_conference_profile.video_grid.name
}
