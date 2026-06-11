package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories wires the in-process provider for acceptance
// tests (no separate install needed).
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"freeswitch": providerserver.NewProtocol6WithError(New("test")()),
}

// providerConfig is prepended to every acceptance test config. It points at the
// local control-plane over HTTPS (skipping verification for the test).
const providerConfig = `
provider "freeswitch" {
  endpoint = "https://localhost:8080"
  token    = "dev-token"
  insecure = true
}
`

func testAccPreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 and run a control-plane to execute acceptance tests")
	}
}
