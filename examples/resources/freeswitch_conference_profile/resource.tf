# A video conference profile: composed grid, 720p. Profiles are read when a
# NEW conference starts, so changes need no reload.
resource "freeswitch_conference_profile" "video_grid" {
  name       = "video-grid"
  video_mode = "mux"

  # defaults: group:grid layout, 1280x720, 15 fps, 48 kHz, MOH while alone
  video_layout      = "group:grid"
  video_canvas_size = "1280x720"
}
