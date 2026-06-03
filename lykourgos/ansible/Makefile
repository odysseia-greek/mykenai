INVENTORY_HELLENISTIKE = inventories/hellenistike/hosts.yaml
INVENTORY_HELLAS = inventories/hellas/hosts.yaml
INVENTORY_SINGLE = inventories/single/hosts.yaml
INVENTORY_ROMAIOI_SINGLE = inventories/romaioi/single/hosts.yaml
INVENTORY_ROMAIOI_HA = inventories/romaioi/ha/hosts.yaml
RASPBERRY_PASSWORD_SECRET ?= odysseia/raspberry/password
RASPBERRY_OLD_PASSWORD ?= raspberry
ANSIBLE_FLAGS ?=

.PHONY: ping ping-prd bootstrap bootstrap-hellas rotate-hellas-password bootstrap-hellenistike bootstrap-prd rook-ceph k0s-hellas k0s-hellenistike k3s-hellenistike lvm-storage-hellas k0s-single k0s-lima-single k0s-lima-ha

KUBECONFIG_FILE ?= $(HOME)/.kube/config
K0S_KUBECONFIG := playbooks/k0s-kubeconfig-byzantion
CLUSTER_NAME := byzantion-odysseia
CONTEXT_NAME := byzantion-odysseia
USER_NAME := byzantion-odysseia-admin

ping-prd:
	ansible -i $(INVENTORY_PRD) all -m ping

ping:
	ansible -i $(INVENTORY_HELLENISTIKE) all -m ping

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

bootstrap-hellas:
	@BECOME_PASS_FILE=$$(mktemp); \
	trap 'rm -f "$$BECOME_PASS_FILE"' EXIT; \
	chmod 600 "$$BECOME_PASS_FILE"; \
	gopass show "$(RASPBERRY_PASSWORD_SECRET)" > "$$BECOME_PASS_FILE"; \
	ANSIBLE_BECOME_PASSWORD_FILE="$$BECOME_PASS_FILE" ansible-playbook $(ANSIBLE_FLAGS) -i $(INVENTORY_HELLAS) playbooks/bootstrap-raspies.yaml

rotate-hellas-password:
	@BECOME_PASS_FILE=$$(mktemp); \
	EXTRA_VARS_FILE=$$(mktemp); \
	trap 'rm -f "$$BECOME_PASS_FILE" "$$EXTRA_VARS_FILE"' EXIT; \
	chmod 600 "$$BECOME_PASS_FILE" "$$EXTRA_VARS_FILE"; \
	printf '%s\n' "$(RASPBERRY_OLD_PASSWORD)" > "$$BECOME_PASS_FILE"; \
	NEW_PASSWORD=$$(gopass show "$(RASPBERRY_PASSWORD_SECRET)"); \
	PASSWORD_HASH=$$(printf '%s\n' "$$NEW_PASSWORD" | openssl passwd -6 -stdin); \
	unset NEW_PASSWORD; \
	printf "raspberry_password_hash: '%s'\n" "$$PASSWORD_HASH" > "$$EXTRA_VARS_FILE"; \
	unset PASSWORD_HASH; \
	ANSIBLE_BECOME_PASSWORD_FILE="$$BECOME_PASS_FILE" ansible-playbook $(ANSIBLE_FLAGS) -i $(INVENTORY_HELLAS) playbooks/rotate-raspberry-password.yaml -e "@$$EXTRA_VARS_FILE"

bootstrap: bootstrap-hellenistike

bootstrap-hellenistike:
	ansible-playbook -i $(INVENTORY_HELLENISTIKE) playbooks/bootstrap-hellenistike.yaml

rook-ceph:
	ansible-playbook -i $(INVENTORY_HELLENISTIKE) playbooks/rook-ceph.yaml

bootstrap-single:
	ansible-playbook -i $(INVENTORY_SINGLE) playbooks/bootstrap-raspies.yaml

k0s-hellenistike:
	ansible-playbook -i $(INVENTORY_HELLENISTIKE) playbooks/k0s-hellenistike.yaml

k0s-hellas:
	@BECOME_PASS_FILE=$$(mktemp); \
	trap 'rm -f "$$BECOME_PASS_FILE"' EXIT; \
	chmod 600 "$$BECOME_PASS_FILE"; \
	gopass show "$(RASPBERRY_PASSWORD_SECRET)" > "$$BECOME_PASS_FILE"; \
	ANSIBLE_BECOME_PASSWORD_FILE="$$BECOME_PASS_FILE" ansible-playbook $(ANSIBLE_FLAGS) -i $(INVENTORY_HELLAS) playbooks/k0s-hellas.yaml

k3s-hellenistike:
	ansible-playbook -i $(INVENTORY_HELLENISTIKE) playbooks/k3s-install-hellenistike.yaml

lvm-storage:
	ansible-playbook -i $(INVENTORY_SINGLE) playbooks/lvm-node-storage.yaml

lvm-storage-hellas:
	@BECOME_PASS_FILE=$$(mktemp); \
	trap 'rm -f "$$BECOME_PASS_FILE"' EXIT; \
	chmod 600 "$$BECOME_PASS_FILE"; \
	gopass show "$(RASPBERRY_PASSWORD_SECRET)" > "$$BECOME_PASS_FILE"; \
	ANSIBLE_BECOME_PASSWORD_FILE="$$BECOME_PASS_FILE" ansible-playbook $(ANSIBLE_FLAGS) -i $(INVENTORY_HELLAS) playbooks/topolvm-prereqs.yaml

k0s-single:
	ansible-playbook -i $(INVENTORY_SINGLE) playbooks/k0s-single-node.yaml

k0s-lima-single: ## Install k0s single-node on Lima VM
	ansible-playbook -i $(INVENTORY_ROMAIOI_SINGLE) playbooks/k0s-lima-single.yaml

k0s-lima-ha: ## Install k0s HA cluster on Lima VMs
	ansible-playbook -i $(INVENTORY_ROMAIOI_HA) playbooks/k0s-lima-ha.yaml

poc: bootstrap-single lvm-storage k0s-single
