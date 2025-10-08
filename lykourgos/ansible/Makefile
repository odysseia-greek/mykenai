
.PHONY: help merge-k0s-config clean-k0s-config switch-k0s view-contexts

KUBECONFIG_FILE ?= $(HOME)/.kube/config
K0S_KUBECONFIG := k0s-kubeconfig-pella
CLUSTER_NAME := acc-odysseia-single
CONTEXT_NAME := acc-odysseia-single
USER_NAME := acc-odysseia-single-admin

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

merge-k0s-config: ## Merge k0s kubeconfig into main kubeconfig with meaningful names
	@echo "Merging k0s kubeconfig into $(KUBECONFIG_FILE)..."
	@if [ ! -f $(K0S_KUBECONFIG) ]; then \
		echo "Error: $(K0S_KUBECONFIG) not found. Run the Ansible playbook first."; \
		exit 1; \
	fi
	@# Backup existing kubeconfig
	@if [ -f $(KUBECONFIG_FILE) ]; then \
		cp $(KUBECONFIG_FILE) $(KUBECONFIG_FILE).backup-$$(date +%Y%m%d-%H%M%S); \
		echo "Created backup of existing kubeconfig"; \
	fi
	@# Create temp file with renamed contexts
	@TMP_CONFIG=$$(mktemp); \
	kubectl config view --kubeconfig=$(K0S_KUBECONFIG) --raw | \
	sed -e 's/name: local/name: $(CLUSTER_NAME)/g' \
	    -e 's/cluster: local/cluster: $(CLUSTER_NAME)/g' \
	    -e 's/name: Default/name: $(CONTEXT_NAME)/g' \
	    -e 's/context: Default/context: $(CONTEXT_NAME)/g' \
	    -e 's/name: user/name: $(USER_NAME)/g' \
	    -e 's/user: user/user: $(USER_NAME)/g' > $$TMP_CONFIG; \
	KUBECONFIG=$(KUBECONFIG_FILE):$$TMP_CONFIG kubectl config view --flatten > $(KUBECONFIG_FILE).new; \
	mv $(KUBECONFIG_FILE).new $(KUBECONFIG_FILE); \
	rm -f $$TMP_CONFIG
	@echo "✓ k0s config merged successfully"
	@echo ""
	@echo "Available contexts:"
	@kubectl config get-contexts
	@echo ""
	@echo "To switch to k0s cluster, run: make switch-k0s"

clean-k0s-config: ## Remove k0s context from kubeconfig
	@echo "Removing k0s context from kubeconfig..."
	@kubectl config delete-context $(CONTEXT_NAME) 2>/dev/null || true
	@kubectl config delete-cluster $(CLUSTER_NAME) 2>/dev/null || true
	@kubectl config delete-user $(USER_NAME) 2>/dev/null || true
	@echo "✓ k0s context removed"

switch-k0s: ## Switch kubectl context to k0s cluster
	@kubectl config use-context $(CONTEXT_NAME)
	@echo "✓ Switched to k0s cluster"
	@echo ""
	@kubectl get nodes

view-contexts: ## View all available kubectl contexts
	@kubectl config get-contexts

deploy-k0s: ## Deploy k0s cluster using Ansible
	@echo "Deploying k0s cluster..."
	@./deploy-k0s.sh
	@echo ""
	@echo "Now run 'make merge-k0s-config' to add the cluster to your kubeconfig"

add-worker-k0s: ## Deploy k0s cluster using Ansible
	@echo "Adding k0s workers..."
	@ansible-playbook -i inventory.ini k0s-join-nodes.yaml
	@echo ""
	@echo "Now run 'make merge-k0s-config' to add the cluster to your kubeconfig"

full-setup: deploy-k0s add-worker-k0s merge-k0s-config ## Deploy k0s and merge config in one step
	@echo "✓ Full k0s setup complete!"
	@echo ""
	@echo "To use the cluster:"
	@echo "  make switch-k0s"