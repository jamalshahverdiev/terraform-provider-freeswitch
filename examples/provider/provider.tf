terraform {
  required_providers {
    freeswitch = {
      source = "local/freeswitch"
    }
  }
}

provider "freeswitch" {
  endpoint     = "https://localhost:8080"
  token        = var.freeswitch_token
  ca_cert_file = "../../deploy/tls/ca.crt"
}
