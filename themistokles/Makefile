# Makefile for setting up Flux and the repository locally for themistokles (core)

# Variables
REPO_URL := https://github.com/odysseia-greek/mykenai.git
BRANCH := feat/add-flux
PATH_IN_REPO := ./themistokles/aer/acc
CLUSTER_NAME := themistokles
NAMESPACE := flux-system

# Check if flux CLI is installed
.PHONY: check-flux
check-flux:
	@if ! command -v flux >/dev/null 2>&1; then \
		echo "Flux CLI not found. Installing..."; \
		curl -s https://fluxcd.io/install.sh | bash; \
	else \
		echo "Flux CLI already installed"; \
	fi

# Bootstrap Flux
.PHONY: bootstrap-flux
bootstrap-flux: check-flux
	@echo "Bootstrapping Flux..."
	@flux install
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
	@echo "  bootstrap-flux    - Bootstrap Flux manually"
	@echo "  flux-status       - Check Flux status"
	@echo "  help              - Show this help message"
