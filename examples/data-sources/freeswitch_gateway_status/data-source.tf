# Runtime status of a Sofia gateway (queried over ESL).
data "freeswitch_gateway_status" "provider_main" {
  profile = "external"
  name    = "provider-main"
}
