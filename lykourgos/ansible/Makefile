INVENTORY_ACC = inventories/acc/hosts.yaml
INVENTORY_PRD = inventories/prd/hosts.yaml
INVENTORY_SINGLE = inventories/single/hosts.yaml

.PHONY: ping-acc ping-prd bootstrap-acc bootstrap-prd k0s-acc k3s-acc k0s-single

KUBECONFIG_FILE ?= $(HOME)/.kube/config
K0S_KUBECONFIG := playbooks/k0s-kubeconfig-pella
CLUSTER_NAME := acc-odysseia-single
CONTEXT_NAME := acc-odysseia-single
USER_NAME := acc-odysseia-single-admin

ping-prd:
	ansible -i $(INVENTORY_PRD) all -m ping

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

bootstrap-prd:
	ansible-playbook -i $(INVENTORY_PRD) playbooks/bootstrap-raspies.yaml

bootstrap-single:
	ansible-playbook -i $(INVETORY_SINGLE) playbooks/bootstrap-raspies.yaml

k0s-acc:
	ansible-playbook -i $(INVENTORY_ACC) playbooks/k0s-install-acc.yaml

k3s-acc:
	ansible-playbook -i $(INVENTORY_ACC) playbooks/k3s-install-acc.yaml

k0s-single:
	ansible-playbook -i $(INVENTORY_SINGLE) playbooks/k0s-install-single.yaml