#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Versions
CILIUM_VERSION="${CILIUM_VERSION:-1.18.1}"

# Kubeconfig
KUBECONFIG_FILE="${KUBECONFIG:-$HOME/.kube/config}"
CONTEXT="${KUBE_CONTEXT:-acc-odysseia-single}"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}k0s Cluster Bootstrap Script${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Configuration:"
echo "  Cilium version: ${CILIUM_VERSION}"
echo "  Context: ${CONTEXT}"
echo ""

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed${NC}"
    exit 1
fi

# Check if cilium CLI is available
if ! command -v cilium &> /dev/null; then
    echo -e "${YELLOW}Warning: cilium CLI is not installed${NC}"
    echo "Install it with: brew install cilium-cli"
    echo "Or visit: https://docs.cilium.io/en/stable/gettingstarted/k8s-install-default/#install-the-cilium-cli"
    exit 1
fi

# Check if helm is available
if ! command -v helm &> /dev/null; then
    echo -e "${RED}Error: helm is not installed${NC}"
    echo "Install it with: brew install helm"
    exit 1
fi

# Switch to k0s context
echo -e "${GREEN} Switching to ctx: ${CONTEXT}...${NC}"
kubectl config use-context ${CONTEXT}
echo ""



# Install Cilium
echo -e "${GREEN} Installing Cilium ${CILIUM_VERSION} in cilium namespace...${NC}"

# Create cilium namespace
kubectl create namespace cilium --dry-run=client -o yaml | kubectl apply -f -

cilium install --version ${CILIUM_VERSION} \
    --namespace cilium \
    --set ipam.mode=kubernetes \
    --set kubeProxyReplacement=false \
    --set enableHostFirewall=false \
    --wait

echo ""
echo "Waiting for Cilium to be ready..."
cilium status --wait --namespace cilium

echo ""
echo "Enabling Cilium Hubble..."
cilium hubble enable --ui --namespace cilium

echo ""
echo "Waiting for Hubble to be ready..."
kubectl wait --for=condition=Ready pods -n cilium -l k8s-app=hubble-relay --timeout=300s || true
kubectl wait --for=condition=Ready pods -n cilium -l k8s-app=hubble-ui --timeout=300s || true

echo -e "${GREEN}✓ Cilium installed successfully${NC}"
echo ""


# Summary
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Bootstrap Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Installed components:"
echo "  ✓ Cilium ${CILIUM_VERSION} (CNI + Network Policy + Hubble) - namespace: cilium"
echo ""
echo "Useful commands:"
echo "  - Check Cilium status:   cilium status -n cilium"
echo "  - Port-forward Hubble:   cilium hubble port-forward -n cilium"
echo "  - Open Hubble UI:        cilium hubble ui -n cilium"
