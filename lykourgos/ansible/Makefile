INVENTORY_ACC = inventories/acc/hosts.yaml
INVENTORY_PRD = inventories/prd/hosts.yaml
INVENTORY_SINGLE = inventories/single/hosts.yaml

.PHONY: ping-acc ping-prd bootstrap-acc bootstrap-prd k0s-acc

ping-acc:
	ansible -i $(INVENTORY_ACC) all -m ping

ping-prd:
	ansible -i $(INVENTORY_PRD) all -m ping

bootstrap-acc:
	ansible-playbook -i $(INVENTORY_ACC) playbooks/bootstrap-raspies.yaml

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