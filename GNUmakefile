HOSTNAME = registry.terraform.io
NAMESPACE = local
NAME      = freeswitch
BINARY    = terraform-provider-$(NAME)
VERSION   = 0.1.0
OS_ARCH  ?= linux_amd64
PLUGINDIR = $(HOME)/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)

default: install

build:
	go build -o $(BINARY) .

# Install into GOBIN (use with a dev_overrides block in ~/.terraformrc).
install: build
	go install .

# Install into the CLI filesystem mirror (use with required_providers + terraform init).
install-mirror: build
	mkdir -p $(PLUGINDIR)
	cp $(BINARY) $(PLUGINDIR)/

fmt:
	gofmt -s -w -e .

vet:
	go vet ./...

test:
	go test ./... -count=1

# Acceptance tests run real terraform against a live control-plane.
# Requires the stack up (docker compose) and a terraform/tofu binary.
testacc:
	TF_ACC=1 go test ./... -v -count=1 -timeout 120m

# Generate provider docs from schema + examples/ into docs/.
# Run tfplugindocs on demand (not pinned in go.mod to avoid dep conflicts).
generate:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest generate --provider-name freeswitch

.PHONY: default build install install-mirror fmt vet test testacc generate
