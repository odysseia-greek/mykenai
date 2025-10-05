
#!/bin/bash

set -e

echo "Deploying k0s single node cluster..."

# Run the k0s playbook
ansible-playbook -i inventory.ini k0s-single-node.yaml

echo "k0s deployment completed!"
echo ""
echo "To use the cluster:"
echo "1. Copy the kubeconfig: cp k0s-kubeconfig-k0s-constantinopel ~/.kube/config"
echo "2. Or set KUBECONFIG: export KUBECONFIG=./k0s-kubeconfig-k0s-constantinopel"
echo "3. Test with: kubectl get nodes"