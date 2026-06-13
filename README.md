# terraform-provider-freeswitch

Terraform provider for the [FreeSWITCH IaC
platform](https://github.com/jamalshahverdiev/freeswitch-iac-platform)
control-plane, built with the
[Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).

Manage FreeSWITCH declaratively: users, dialplans/IVRs, call-center queues,
audio/video conference rooms — all stored in PostgreSQL and served to
FreeSWITCH via `mod_xml_curl`.

> ⚠️ **This provider requires the FreeSWITCH IaC control-plane.** It is a
> Terraform client for that API and does nothing on its own — every resource
> is a call to the control-plane, which must be running and reachable. It
> never talks to FreeSWITCH (ESL/SIP) directly. Deploy the control-plane
> first: **https://github.com/jamalshahverdiev/freeswitch-iac-platform**
>
> ```
> Terraform ─► terraform-provider-freeswitch ─► Control-Plane API ─► PostgreSQL
>                                                      │  (mod_xml_curl / ESL)
>                                                      ▼
>                                                 FreeSWITCH
> ```

## Resources

- `freeswitch_domain` — directory domain (id = `name`)
- `freeswitch_user` — SIP user / extension (id = `domain/number`)
- `freeswitch_gateway` — Sofia gateway / trunk (id = `profile/name`)
- `freeswitch_dialplan_extension` — dialplan extension with nested
  `condition { action {} }` blocks (id = `<uuid>`). **IVRs are built from
  these** (`answer` + `play_and_get_digits` + `transfer` + routing extensions).
- `freeswitch_callcenter_queue` / `_agent` / `_tier` — mod_callcenter
  desired state (queue id = `name`, tier id = `queue/agent`)
- `freeswitch_conference_profile` / `_room` — mod_conference profiles and
  rooms; a room also materializes its dialplan entry extension
- `freeswitch_reloadxml` — runs FreeSWITCH `reloadxml` via ESL when its
  `triggers` change
- `freeswitch_callcenter_reload` — runs `reload mod_callcenter` (the apply
  step for queue/agent/tier changes; key `triggers` on `updated_at`)

## Data sources

- Config lookups: `freeswitch_domain`, `freeswitch_user`,
  `freeswitch_gateway`, `freeswitch_dialplan_extension`,
  `freeswitch_callcenter_queue` / `_agent` / `_tier`,
  `freeswitch_conference_profile` / `_room`
- Runtime (ESL): `freeswitch_gateway_status`,
  `freeswitch_user_registration`, `freeswitch_conference_status`

Generated reference with examples: [`docs/`](docs/).

## Provider configuration

```hcl
provider "freeswitch" {
  endpoint     = "https://localhost:8080" # or FREESWITCH_ENDPOINT
  token        = var.token                # or FREESWITCH_TOKEN
  ca_cert_file = "/path/to/ca.crt"        # or FREESWITCH_CACERT
  # insecure   = true                     # or FREESWITCH_INSECURE=1 (skip TLS verify)
}
```

End-to-end usage examples (IVRs, WebRTC users, call center, video
conferences) live in the platform repo:
[freeswitch-iac-platform/examples](https://github.com/jamalshahverdiev/freeswitch-iac-platform/tree/main/examples).

## Local development (dev_overrides — no `terraform init`)

```bash
make install            # go install . -> $GOBIN
```

Add to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/local/freeswitch" = "<output of: go env GOBIN || echo $(go env GOPATH)/bin>"
    # for OpenTofu also add:
    # "registry.opentofu.org/local/freeswitch" = "<same path>"
  }
  direct {}
}
```

Then `terraform plan` / `apply` use the local build directly (skip `init`).

Alternative (filesystem mirror, requires `terraform init`):

```bash
make install-mirror
```

## Tests

```bash
make test                       # unit
make testacc                    # acceptance (needs a running control-plane)
# with OpenTofu:
TF_ACC_TERRAFORM_PATH=$(command -v tofu) make testacc
```

## Docs

```bash
make generate   # tfplugindocs generate (schema + examples/ -> docs/)
```

## Import

```bash
terraform import freeswitch_domain.main 192.168.48.143
terraform import freeswitch_user.u2001 192.168.48.143/2001
terraform import freeswitch_gateway.gw external/provider-main
terraform import freeswitch_dialplan_extension.ext <uuid>
terraform import freeswitch_callcenter_queue.support "support@192.168.48.143"
terraform import freeswitch_callcenter_tier.t "support@192.168.48.143/4201@192.168.48.143"
terraform import freeswitch_conference_room.standup standup
```
