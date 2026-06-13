terraform {
  required_providers {
    freeswitch = {
      source  = "jamalshahverdiev/freeswitch"
      version = "~> 0.1"
    }
  }
}

# Points at the FreeSWITCH IaC control-plane API (NOT at FreeSWITCH itself).
# Run the control-plane first — see
# https://github.com/jamalshahverdiev/freeswitch-iac-platform
provider "freeswitch" {
  endpoint     = "https://localhost:8080" # control-plane base URL
  token        = var.freeswitch_token     # bearer token for /api/v1
  ca_cert_file = "ca.crt"                 # CA that signed the control-plane cert
}
