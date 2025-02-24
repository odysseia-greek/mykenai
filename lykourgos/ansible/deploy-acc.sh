#!/bin/bash

set -e  # Exit immediately if a command fails
set -u  # Treat unset variables as an error

INVENTORY="inventory.ini"

echo "🚀 Running bootstrap setup for ACC cluster..."
ansible-playbook -i $INVENTORY bootstrap-raspies.yaml --limit poleis-acc

echo "⏳ Installing K3s on ACC cluster..."
ansible-playbook -i $INVENTORY k3s-acc.yaml --limit poleis-acc

echo "🔧 Running post-install configuration (kubeconfig merge)..."
ansible-playbook -i $INVENTORY post-install.yaml --limit servers-acc

echo "✅ ACC Cluster setup completed!"
kubectl config get-contexts
