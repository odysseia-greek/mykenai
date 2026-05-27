.PHONY: help single-start single-stop single-delete ha-start ha-stop ha-delete kubeconfig-single kubeconfig-ha status clean create-disk-single create-disk-ha update-hosts

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
	@if limactl disk ls 2>/dev/null | grep -q "pyxis-byzantion"; then \
		echo "Disk 'pyxis-byzantion' already exists"; \
	else \
		limactl disk create pyxis-byzantion --size 30G; \
		echo "Disk 'pyxis-byzantion' created (30GB)"; \
	fi
	@if limactl disk ls 2>/dev/null | grep -q "pyxis-trapezous"; then \
		echo "Disk 'pyxis-trapezous' already exists"; \
	else \
		limactl disk create pyxis-trapezous --size 30G; \
		echo "Disk 'pyxis-trapezous' created (30GB)"; \
	fi
	@if limactl disk ls 2>/dev/null | grep -q "pyxis-nikaia"; then \
		echo "Disk 'pyxis-nikaia' already exists"; \
	else \
		limactl disk create pyxis-nikaia --size 30G; \
		echo "Disk 'pyxis-nikaia' created (30GB)"; \
	fi

list-disks:
	@echo "Lima disks:"
	@limactl disk ls 2>/dev/null | grep -E "NAME|pyxis" || echo "No pyxis disks found"

update-hosts:
	@echo "Updating /etc/hosts for byzantion..."
	@grep -q "byzantion.odysseia-greek" /etc/hosts || \
		echo "192.168.105.2 byzantion.odysseia-greek" | sudo tee -a /etc/hosts > /dev/null
	@echo "Hosts entry present: 192.168.105.2 byzantion.odysseia-greek"

delete-disks:
	@echo "Deleting k0s disks..."
	-limactl disk delete pyxis 2>/dev/null
	-limactl disk delete pyxis-byzantion 2>/dev/null
	-limactl disk delete pyxis-trapezous 2>/dev/null
	-limactl disk delete pyxis-nikaia 2>/dev/null
	@echo "Disks deleted"

# Single-node cluster
single-start: create-disk-single
	@echo "Starting k0s single-node cluster..."
	limactl start --yes --name=byzantion k0s-single.yaml
	@echo ""
	@echo "Cluster started! Get kubeconfig with: make kubeconfig-single"
	@echo "Don't forget to add to /etc/hosts:"
	@echo "  192.168.105.2 byzantion.odysseia-greek"

single-stop:
	limactl stop byzantion

single-delete:
	limactl delete byzantion
	@echo "VM deleted. Disk 'pyxis' preserved. To delete disk: make delete-disks"

single-restart: single-stop single-delete delete-disks single-start

kubeconfig-single:
	@mkdir -p ~/.kube
	limactl shell byzantion sudo k0s kubeconfig admin > ~/.kube/byzantion-config
	@echo "Kubeconfig saved to: ~/.kube/byzantion-config"
	@echo "Export with: export KUBECONFIG=~/.kube/byzantion-config"

# HA cluster
ha-start: create-disk-ha
	@echo "Starting k0s HA cluster..."
	@echo "Starting controller..."
	limactl start --yes --name=byzantion k0s-ha.yaml
	@echo "Starting worker 1..."
	limactl start --yes --name=trapezous --set='.additionalDisks[0].name="pyxis-trapezous"' k0s-ha-worker.yaml
	@echo "Starting worker 2..."
	limactl start --yes --name=nikaia --set='.additionalDisks[0].name="pyxis-nikaia"' k0s-ha-worker.yaml
	@echo "Installing k0s with Ansible..."
	cd ../ansible && ansible-playbook -i inventories/romaioi/ha/hosts.yaml playbooks/k0s-lima-ha.yaml
	@echo ""
	@echo "HA cluster started! Get kubeconfig with: make kubeconfig-ha"
	@echo "Don't forget to add to /etc/hosts:"
	@echo "  192.168.105.2 byzantion.odysseia-greek"

ha-stop:
	-limactl stop byzantion
	-limactl stop trapezous
	-limactl stop nikaia

ha-delete:
	-limactl delete byzantion
	-limactl delete trapezous
	-limactl delete nikaia
	@echo "VMs deleted. Disks preserved. To delete disks: make delete-disks"

kubeconfig-ha:
	@mkdir -p ~/.kube
	limactl shell byzantion sudo k0s kubeconfig admin > ~/.kube/byzantion-ha-config
	@echo "Kubeconfig saved to: ~/.kube/byzantion-ha-config"
	@echo "Export with: export KUBECONFIG=~/.kube/byzantion-ha-config"

# General commands
status:
	@echo "Lima VMs:"
	@limactl list | grep -E "NAME|byzantion|trapezous|nikaia" || echo "No Romaioi VMs running"

clean:
	@echo "Deleting all k0s VMs and disks..."
	-limactl stop byzantion 2>/dev/null
	-limactl stop trapezous 2>/dev/null
	-limactl stop nikaia 2>/dev/null
	-limactl delete byzantion 2>/dev/null
	-limactl delete trapezous 2>/dev/null
	-limactl delete nikaia 2>/dev/null
	$(MAKE) delete-disks
	@echo "Cleanup complete"

# Node operations
shell-single:
	limactl shell byzantion

shell-controller:
	limactl shell byzantion

shell-worker1:
	limactl shell trapezous

shell-worker2:
	limactl shell nikaia

logs-single:
	limactl shell byzantion sudo journalctl -u k0scontroller -f

logs-controller:
	limactl shell byzantion sudo journalctl -u k0scontroller -f

logs-worker1:
	limactl shell trapezous sudo journalctl -u k0sworker -f

logs-worker2:
	limactl shell nikaia sudo journalctl -u k0sworker -f
