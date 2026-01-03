.PHONY: help single-start single-stop single-delete ha-start ha-stop ha-delete kubeconfig-single kubeconfig-ha status clean create-disk-single create-disk-ha

help:
	@echo "k0s Lima Cluster Management"
	@echo ""
	@echo "Single-node cluster:"
	@echo "  make single-start       - Start single-node cluster (creates disk)"
	@echo "  make single-stop        - Stop single-node cluster"
	@echo "  make single-delete      - Delete single-node cluster (keeps disk)"
	@echo "  make kubeconfig-single  - Get kubeconfig for single-node"
	@echo ""
	@echo "HA cluster:"
	@echo "  make ha-start           - Start HA cluster (controller + 2 workers, creates disks)"
	@echo "  make ha-stop            - Stop HA cluster"
	@echo "  make ha-delete          - Delete HA cluster (keeps disks)"
	@echo "  make kubeconfig-ha      - Get kubeconfig for HA cluster"
	@echo ""
	@echo "Disk management:"
	@echo "  make create-disk-single - Create 30GB disk for single-node"
	@echo "  make create-disk-ha     - Create 30GB disks for HA cluster"
	@echo "  make list-disks         - List all Lima disks"
	@echo "  make delete-disks       - Delete all k0s disks"
	@echo ""
	@echo "General:"
	@echo "  make status             - Show status of all VMs"
	@echo "  make clean              - Delete all k0s VMs and disks"
	@echo "  make update-hosts       - Update /etc/hosts entries"

# Disk management
create-disk-single:
	@echo "Creating 30GB disk for single-node cluster..."
	@if limactl disk ls 2>/dev/null | grep -q "pyxis"; then \
		echo "Disk 'pyxis' already exists"; \
	else \
		limactl disk create pyxis --size 30G; \
		echo "Disk 'pyxis' created (30GB)"; \
	fi

create-disk-ha:
	@echo "Creating 30GB disks for HA cluster..."
	@if limactl disk ls 2>/dev/null | grep -q "pyxis-controller"; then \
		echo "Disk 'pyxis-controller' already exists"; \
	else \
		limactl disk create pyxis-controller --size 30G; \
		echo "Disk 'pyxis-controller' created (30GB)"; \
	fi
	@if limactl disk ls 2>/dev/null | grep -q "pyxis-worker1"; then \
		echo "Disk 'pyxis-worker1' already exists"; \
	else \
		limactl disk create pyxis-worker1 --size 30G; \
		echo "Disk 'pyxis-worker1' created (30GB)"; \
	fi
	@if limactl disk ls 2>/dev/null | grep -q "pyxis-worker2"; then \
		echo "Disk 'pyxis-worker2' already exists"; \
	else \
		limactl disk create pyxis-worker2 --size 30G; \
		echo "Disk 'pyxis-worker2' created (30GB)"; \
	fi

list-disks:
	@echo "Lima disks:"
	@limactl disk ls 2>/dev/null | grep -E "NAME|pyxis" || echo "No pyxis disks found"

delete-disks:
	@echo "Deleting k0s disks..."
	-limactl disk delete pyxis 2>/dev/null
	-limactl disk delete pyxis-controller 2>/dev/null
	-limactl disk delete pyxis-worker1 2>/dev/null
	-limactl disk delete pyxis-worker2 2>/dev/null
	@echo "Disks deleted"

# Single-node cluster
single-start: create-disk-single
	@echo "Starting k0s single-node cluster..."
	limactl start --yes --name=k0s-byzantium k0s-single.yaml
	@echo ""
	@echo "Cluster started! Get kubeconfig with: make kubeconfig-single"
	@echo "Don't forget to add to /etc/hosts:"
	@echo "  192.168.105.2 k0s-byzantium.odysseia-greek byzantium.odysseia-greek"

single-stop:
	limactl stop k0s-byzantium

single-delete:
	limactl delete k0s-byzantium
	@echo "VM deleted. Disk 'pyxis' preserved. To delete disk: make delete-disks"

single-restart: single-stop single-delete delete-disks single-start

kubeconfig-single:
	@mkdir -p ~/.kube
	limactl shell k0s-byzantium sudo k0s kubeconfig admin > ~/.kube/k0s-byzantium-config
	@echo "Kubeconfig saved to: ~/.kube/k0s-byzantium-config"
	@echo "Export with: export KUBECONFIG=~/.kube/k0s-byzantium-config"

# HA cluster
ha-start: create-disk-ha
	@echo "Starting k0s HA cluster..."
	@echo "Starting controller..."
	limactl start --name=k0s-controller k0s-ha.yaml
	@echo "Waiting for controller to be ready..."
	@sleep 30
	@echo "Getting worker token..."
	$(eval TOKEN := $(shell limactl shell k0s-controller cat /tmp/worker-token.txt))
	@echo "Starting worker 1..."
	limactl start --name=k0s-worker1 --set='.networks[0].macAddress="52:55:55:12:34:61"' --set='.additionalDisks[0].name="pyxis-worker1"' k0s-ha-worker.yaml
	@echo "Joining worker 1..."
	@echo "$(TOKEN)" | limactl shell k0s-worker1 sudo tee /tmp/worker-token.txt > /dev/null
	limactl shell k0s-worker1 sudo k0s install worker --token-file /tmp/worker-token.txt
	limactl shell k0s-worker1 sudo systemctl start k0sworker
	limactl shell k0s-worker1 sudo systemctl enable k0sworker
	@echo "Starting worker 2..."
	limactl start --name=k0s-worker2 --set='.networks[0].macAddress="52:55:55:12:34:62"' --set='.additionalDisks[0].name="pyxis-worker2"' k0s-ha-worker.yaml
	@echo "Joining worker 2..."
	@echo "$(TOKEN)" | limactl shell k0s-worker2 sudo tee /tmp/worker-token.txt > /dev/null
	limactl shell k0s-worker2 sudo k0s install worker --token-file /tmp/worker-token.txt
	limactl shell k0s-worker2 sudo systemctl start k0sworker
	limactl shell k0s-worker2 sudo systemctl enable k0sworker
	@echo ""
	@echo "HA cluster started! Get kubeconfig with: make kubeconfig-ha"
	@echo "Don't forget to add to /etc/hosts:"
	@echo "  192.168.105.10 k0s-byzantium.odysseia-greek byzantium.odysseia-greek"

ha-stop:
	limactl stop k0s-controller k0s-worker1 k0s-worker2

ha-delete:
	limactl delete k0s-controller k0s-worker1 k0s-worker2
	@echo "VMs deleted. Disks preserved. To delete disks: make delete-disks"

kubeconfig-ha:
	@mkdir -p ~/.kube
	limactl shell k0s-controller sudo k0s kubeconfig admin > ~/.kube/k0s-byzantium-ha-config
	@echo "Kubeconfig saved to: ~/.kube/k0s-byzantium-ha-config"
	@echo "Export with: export KUBECONFIG=~/.kube/k0s-byzantium-ha-config"

# General commands
status:
	@echo "Lima VMs:"
	@limactl list | grep -E "NAME|k0s-" || echo "No k0s VMs running"

clean:
	@echo "Deleting all k0s VMs and disks..."
	-limactl stop k0s-byzantium k0s-controller k0s-worker1 k0s-worker2 2>/dev/null
	-limactl delete k0s-byzantium k0s-controller k0s-worker1 k0s-worker2 2>/dev/null
	$(MAKE) delete-disks
	@echo "Cleanup complete"

update-hosts:
	@echo "Adding k0s entries to /etc/hosts..."
	@grep -q "k0s-byzantium.odysseia-greek" /etc/hosts 2>/dev/null || \
		echo "192.168.105.2 k0s-byzantium.odysseia-greek byzantium.odysseia-greek  # k0s single" | sudo tee -a /etc/hosts
	@grep -q "192.168.105.10" /etc/hosts 2>/dev/null || \
		echo "192.168.105.10 k0s-byzantium.odysseia-greek byzantium.odysseia-greek  # k0s HA" | sudo tee -a /etc/hosts
	@echo "/etc/hosts updated"

# Node operations
shell-single:
	limactl shell k0s-byzantium

shell-controller:
	limactl shell k0s-controller

shell-worker1:
	limactl shell k0s-worker1

shell-worker2:
	limactl shell k0s-worker2

logs-single:
	limactl shell k0s-byzantium sudo journalctl -u k0scontroller -f

logs-controller:
	limactl shell k0s-controller sudo journalctl -u k0scontroller -f

logs-worker1:
	limactl shell k0s-worker1 sudo journalctl -u k0sworker -f

logs-worker2:
	limactl shell k0s-worker2 sudo journalctl -u k0sworker -f
