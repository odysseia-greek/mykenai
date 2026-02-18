# Makefile for setting up Flux and the repository locally for themistokles (core)

# Variables
REPO_URL := https://github.com/odysseia-greek/mykenai.git
BRANCH := main
ENV ?= romaioi
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
	@kubectl get ns $(SOPS_NS) >/dev/null 2>&1 || kubectl create ns $(SOPS_NS)
	@echo "Waiting for namespace $(SOPS_NS) to be ready..."
	@until kubectl get ns $(SOPS_NS) >/dev/null 2>&1; do sleep 1; done

ensure-sops-gpg: ensure-flux-ns
	@echo "Creating/updating SOPS GPG secret '$(SOPS_SECRET)' in namespace '$(SOPS_NS)'..."
	@gpg --batch --yes --pinentry-mode loopback \
	  --export-secret-keys --armor "$(SOPS_KEY_FP)" | \
	  kubectl -n "$(SOPS_NS)" create secret generic "$(SOPS_SECRET)" \
	    --from-file=sops.asc=/dev/stdin \
	    --dry-run=client -o yaml | \
	  kubectl apply -f -
	@echo "Done."

# Bootstrap Flux
.PHONY: bootstrap-flux
bootstrap-flux: check-flux
	@echo "Bootstrapping Flux for environment: $(ENV)..."
	@echo "Using path: $(PATH_IN_REPO)"

	# ensure sops key exists BEFORE any kustomization that needs decryption
	@$(MAKE) ensure-sops-gpg

	@flux create source git themistokles \
	  --url="${REPO_URL}" \
	  --branch="${BRANCH}" \
	  --interval=1m

	# kustomization for core
	@flux create kustomization themistokles \
	  --namespace=flux-system \
	  --source=themistokles \
	  --path="${PATH_IN_REPO}" \
	  --prune=true \
	  --interval=10m

# Check Flux status
.PHONY: flux-status
flux-status:
	@echo "Checking Flux status..."
	@flux get all

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  check-flux        - Check if Flux CLI is installed"
	@echo "  bootstrap-flux    - Bootstrap Flux manually (ENV=romaioi|hellenistike|hellas, default: local)"
	@echo "  flux-status       - Check Flux status"
	@echo "  help              - Show this help message"
	@echo ""
	@echo "Usage examples:"
	@echo "  make bootstrap-flux              # Uses romaioi environment (default)"
	@echo "  make bootstrap-flux ENV=acc      # Uses hellenistike environment"
	@echo "  make bootstrap-flux ENV=hellas     # Uses hellas environment"