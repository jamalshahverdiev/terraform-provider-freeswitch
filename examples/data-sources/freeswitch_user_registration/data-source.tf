# Runtime SIP registration state for a user (queried over ESL).
data "freeswitch_user_registration" "u2001" {
  user   = "2001"
  domain = "192.168.48.143"
}
