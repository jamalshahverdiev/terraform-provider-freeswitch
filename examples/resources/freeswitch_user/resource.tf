resource "freeswitch_user" "u2001" {
  domain = freeswitch_domain.main.name
  number = "2001"

  params = {
    password    = var.user_password
    vm-password = "2001"
  }

  variables = {
    effective_caller_id_name   = "User 2001"
    effective_caller_id_number = "2001"
    user_context               = "company"
  }
}
