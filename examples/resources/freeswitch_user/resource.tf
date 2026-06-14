resource "freeswitch_user" "u2001" {
  domain = freeswitch_domain.main.name
  number = "2001"

  params = {
    password = var.user_password
  }

  variables = {
    effective_caller_id_name   = "User 2001"
    effective_caller_id_number = "2001"
    user_context               = "company"
  }

  # Typed mailbox — rendered into the directory as vm-* params. Prefer this
  # over setting vm-password/vm-mailto in params by hand.
  voicemail = {
    password    = "2001"
    email       = "user2001@example.com"
    attach_file = true
  }
}
