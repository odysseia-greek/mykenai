#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Versions
CILIUM_VERSION="${CILIUM_VERSION:-1.18.1}"
FLUX_NAMESPACE="${FLUX_NAMESPACE:-flux-system}"

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

# Install Cilium
echo -e "${GREEN} Installing Cilium ${CILIUM_VERSION} in cilium namespace...${NC}"

# Create cilium namespace
kubectl create namespace cilium --dry-run=client -o yaml | kubectl apply -f -

cilium install \
  --version ${CILIUM_VERSION} \
  --namespace cilium \
  --set ipam.mode=kubernetes \
  --set kubeProxyReplacement=false \
  --set enableHostFirewall=false \
  --set envoy.enabled=false \
  --wait

echo "Waiting for cilium namespace to exist..."
for i in {1..60}; do
  kubectl get namespace cilium >/dev/null 2>&1 && break
  sleep 2
done

kubectl get namespace cilium >/dev/null 2>&1 || {
  echo "ERROR: cilium namespace did not appear in time"
  exit 1
}

kubens cilium

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

# ---------- Flux ----------
echo -e "${GREEN}Installing Flux controllers in ${FLUX_NAMESPACE}...${NC}"

# Check flux CLI
if ! command -v flux &> /dev/null; then
  echo -e "${RED}Error: flux CLI is not installed${NC}"
  echo "Install with: brew install fluxcd/tap/flux"
  exit 1
fi

# Namespace
kubectl create namespace "${FLUX_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

# Install controllers (idempotent)
flux install --namespace "${FLUX_NAMESPACE}"

# Wait for core controllers to be ready
kubectl -n "${FLUX_NAMESPACE}" rollout status deploy/source-controller --timeout=180s || true
kubectl -n "${FLUX_NAMESPACE}" rollout status deploy/kustomize-controller --timeout=180s || true
kubectl -n "${FLUX_NAMESPACE}" rollout status deploy/notification-controller --timeout=180s || true


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
