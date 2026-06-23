# Binds a Keycloak identity to a SIP extension. The webphone BFF looks up the
# logged-in user's `sub` here to resolve which extension they register as.
# RBAC roles (agent/supervisor/admin) live in Keycloak, not in this resource.

resource "freeswitch_user" "reception" {
  domain   = "example.com"
  number   = "1001"
  password = var.reception_sip_password
}

resource "freeswitch_operator" "reception" {
  subject      = "a1b2c3d4-0000-1111-2222-333344445555" # Keycloak user `sub`
  domain       = freeswitch_user.reception.domain
  number       = freeswitch_user.reception.number
  display_name = "Reception"
}
