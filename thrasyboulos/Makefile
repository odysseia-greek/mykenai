# Makefile for setting up Flux and the repository locally

# Variables
REPO_URL := https://github.com/odysseia-greek/mykenai.git
BRANCH := feat/add-flux
PATH_IN_REPO := ./thrasyboulos/aer/acc
CLUSTER_NAME := odysseia
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
	@flux create source git thrasyboulos \
		--url=$(REPO_URL) \
		--branch=$(BRANCH) \
		--interval=1m
	@flux create kustomization thrasyboulos\
		--source=thrasyboulos \
		--path=$(PATH_IN_REPO) \
		--prune=true \
		--interval=10m

# Alternative bootstrap method using flux bootstrap command
.PHONY: bootstrap-flux-git
bootstrap-flux-git: check-flux
	@echo "Bootstrapping Flux with Git repository..."
	@flux bootstrap git \
		--url=$(REPO_URL) \
		--branch=$(BRANCH) \
		--path=$(PATH_IN_REPO) \
		--interval=1m

# Check Flux status
.PHONY: flux-status
flux-status:
	@echo "Checking Flux status..."
	@flux get all

# Default target
.PHONY: all
all: bootstrap-flux

# Copy Elasticsearch secret from aristoteles to odysseia namespace
.PHONY: copy-es-secret
copy-es-secret:
	@echo "Copying aristoteles-es-elastic-user secret from aristoteles to odysseia namespace..."
	@kubectl get secret aristoteles-es-elastic-user -n agora -o jsonpath='{.data.elastic}' | \
		awk 'BEGIN {print "apiVersion: v1\ndata:\n  elastic: {print $0}"}  END {print "kind: Secret\nmetadata:\n  name: aristoteles-es-elastic-user\n  namespace: delphi\ntype: Opaque"}' > secret.yaml

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all               - Set up everything (default)"
	@echo "  check-flux        - Check if Flux CLI is installed"
	@echo "  bootstrap-flux    - Bootstrap Flux manually"
	@echo "  bootstrap-flux-git - Bootstrap Flux using the git command"
	@echo "  flux-status       - Check Flux status"
	@echo "  copy-es-secret    - Copy aristoteles-es-elastic-user secret from aristoteles to odysseia namespace"
	@echo "  help              - Show this help message"
