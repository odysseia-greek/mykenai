# Makefile for setting up Flux and the repository locally

# Variables
REPO_URL := https://github.com/odysseia-greek/mykenai.git
BRANCH := feat/add-flux
ENV ?= local
PATH_IN_REPO := ./thrasyboulos/aer/$(ENV)
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

# Default target
.PHONY: all
all: bootstrap-flux

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  check-flux        - Check if Flux CLI is installed"
	@echo "  bootstrap-flux    - Bootstrap Flux manually (ENV=local|acc|prod, default: local)"
	@echo "  flux-status       - Check Flux status"
	@echo "  help              - Show this help message"
	@echo ""
	@echo "Usage examples:"
	@echo "  make bootstrap-flux              # Uses local environment (default)"
	@echo "  make bootstrap-flux ENV=acc      # Uses acc environment"
	@echo "  make bootstrap-flux ENV=prod     # Uses prod environment"