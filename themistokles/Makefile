SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# Makefile for setting up Flux and the repository locally for themistokles (core)

# Variables
REPO_URL := https://github.com/odysseia-greek/mykenai.git
BRANCH ?= main
ENV ?= romaioi
HELLENISTIKE_ENV := hellenistike
HELLAS_ENV := hellas
MINOAN_ENV := minoan
KUBE_CONTEXT ?=
USE_SOPS ?= true
FLUX = flux $(if $(KUBE_CONTEXT),--context "$(KUBE_CONTEXT)")
KUBECTL = kubectl $(if $(KUBE_CONTEXT),--context "$(KUBE_CONTEXT)")
PATH_IN_REPO := ./themistokles/aer/$(ENV)
NAMESPACE := flux-system
SOPS_NS        ?= flux-system
SOPS_SECRET    ?= sops-gpg
SOPS_KEY_FP    ?= 3C8B5BB6281C34C5E80C473086FCFB28CF0EC482

# Check if flux CLI is installed
.PHONY: check-flux
check-flux:
	@if ! command -v flux >/dev/null 2>&1; then \
		echo "Flux CLI not found. Installing..."; \
		curl -s https://fluxcd.io/install.sh | bash; \
	else \
		echo "Flux CLI already installed"; \
	fi

.PHONY: ensure-flux-ns ensure-sops-gpg
ensure-flux-ns:
	@echo "Ensuring namespace $(SOPS_NS) exists..."
	@$(KUBECTL) get ns $(SOPS_NS) >/dev/null 2>&1 || $(KUBECTL) create ns $(SOPS_NS)
	@echo "Waiting for namespace $(SOPS_NS) to be ready..."
	@until $(KUBECTL) get ns $(SOPS_NS) >/dev/null 2>&1; do sleep 1; done

ensure-sops-gpg: ensure-flux-ns
	@echo "Creating/updating SOPS GPG secret '$(SOPS_SECRET)' in namespace '$(SOPS_NS)'..."
	@key_file=$$(mktemp); \
	  trap 'rm -f "$$key_file"' EXIT; \
	  chmod 600 "$$key_file"; \
	  gpg --batch --yes --pinentry-mode loopback \
	    --output "$$key_file" --export-secret-keys --armor "$(SOPS_KEY_FP)"; \
	  test -s "$$key_file" || { echo "No private GPG key exported for $(SOPS_KEY_FP)" >&2; exit 1; }; \
	  $(KUBECTL) -n "$(SOPS_NS)" create secret generic "$(SOPS_SECRET)" \
	    --from-file=sops.asc="$$key_file" \
	    --dry-run=client -o yaml | $(KUBECTL) apply -f -
	@echo "Done."

# Bootstrap Flux
.PHONY: bootstrap-flux bootstrap-hellenistike bootstrap-hellas
bootstrap-flux: check-flux
	@echo "Bootstrapping Flux for environment: $(ENV)..."
	@echo "Using path: $(PATH_IN_REPO)"

	# ensure sops key exists BEFORE any kustomization that needs decryption
	@if [ "$(USE_SOPS)" = true ]; then \
	  $(MAKE) ensure-sops-gpg; \
	fi

	@$(FLUX) create source git themistokles \
	  --url="${REPO_URL}" \
	  --branch="${BRANCH}" \
	  --interval=1m

	# kustomization for core
	@$(FLUX) create kustomization themistokles \
	  --namespace=flux-system \
	  --source=themistokles \
	  --path="${PATH_IN_REPO}" \
	  --prune=true \
	  --interval=10m

bootstrap-hellenistike:
	@$(MAKE) bootstrap-flux ENV=$(HELLENISTIKE_ENV)

bootstrap-hellas:
	@$(MAKE) bootstrap-flux ENV=$(HELLAS_ENV)

# Generate the encrypted Eleusinian test secret with the same bootstrap recipient.
.PHONY: encrypt-minoan-eleusinian
encrypt-minoan-eleusinian:
	@SOPS_KEY_FP="$(SOPS_KEY_FP)" bash hydor/overlays/minoan/eleusinian/create.sh

# k0s and bounded host storage are provisioned separately by Lykourgos Ansible.
.PHONY: bootstrap-minoan
bootstrap-minoan:
	@command -v kubectl >/dev/null && command -v helm >/dev/null && command -v cilium >/dev/null && command -v flux >/dev/null || { echo "Install kubectl, helm, cilium and flux first."; exit 1; }
	@KUBE_CONTEXT="$(if $(KUBE_CONTEXT),$(KUBE_CONTEXT),minoan-odysseia)" \
	  CILIUM_VERSION=1.20.2 \
	  CILIUM_VALUES_FILE="$(CURDIR)/ge/overlays/minoan/cilium/bootstrap-values.yaml" \
	  bash ../lykourgos/k0s/bootstrap.sh
	@$(MAKE) bootstrap-flux ENV=$(MINOAN_ENV) USE_SOPS=true \
	  KUBE_CONTEXT="$(if $(KUBE_CONTEXT),$(KUBE_CONTEXT),minoan-odysseia)"
	@flux --context "$(if $(KUBE_CONTEXT),$(KUBE_CONTEXT),minoan-odysseia)" reconcile source git themistokles
	@flux --context "$(if $(KUBE_CONTEXT),$(KUBE_CONTEXT),minoan-odysseia)" reconcile kustomization themistokles --with-source
	@flux --context "$(if $(KUBE_CONTEXT),$(KUBE_CONTEXT),minoan-odysseia)" get all

# Check Flux status
.PHONY: flux-status
flux-status:
	@echo "Checking Flux status..."
	@$(FLUX) get all

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  check-flux        - Check if Flux CLI is installed"
	@echo "  bootstrap-flux    - Bootstrap Flux manually (ENV=romaioi|hellenistike|hellas|minoan, default: romaioi)"
	@echo "  bootstrap-hellenistike - Bootstrap Flux with the Hellenistike core stack"
	@echo "  bootstrap-hellas  - Bootstrap Flux with the Hellas core stack"
	@echo "  bootstrap-minoan - Bootstrap Cilium, Flux and the Minoan core stack (with SOPS GPG key)"
	@echo "  flux-status       - Check Flux status"
	@echo "  help              - Show this help message"
	@echo ""
	@echo "Usage examples:"
	@echo "  make bootstrap-flux              # Uses romaioi environment (default)"
	@echo "  make bootstrap-hellenistike      # Uses hellenistike environment"
	@echo "  make bootstrap-hellas            # Uses hellas environment"
	@echo "  make bootstrap-minoan BRANCH=main # Use a pushed branch containing Minoan manifests"
	@echo "  make bootstrap-flux ENV=hellas   # Uses hellas environment"
