#!/bin/bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Versions
CILIUM_VERSION="${CILIUM_VERSION:-1.20.0}"
FLUX_NAMESPACE="${FLUX_NAMESPACE:-flux-system}"

# Kubeconfig
export KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config}"
CONTEXT="${KUBE_CONTEXT:-}"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}k0s Cluster Bootstrap Script${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Configuration:"
echo "  Cilium version: ${CILIUM_VERSION}"
echo "  Context: ${CONTEXT:-current kubeconfig context}"
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

# Check flux CLI
if ! command -v flux &> /dev/null; then
  echo -e "${RED}Error: flux CLI is not installed${NC}"
  echo "Install with: brew install fluxcd/tap/flux"
  exit 1
fi

# Resolve one context and use it for every client without changing kubeconfig.
CONTEXT="${CONTEXT:-$(kubectl config current-context)}"
kubectl() { command kubectl --context "$CONTEXT" "$@"; }
helm() { command helm --kube-context "$CONTEXT" "$@"; }
cilium() { command cilium --context "$CONTEXT" "$@"; }
flux() { command flux --context "$CONTEXT" "$@"; }
echo "Using Kubernetes context: ${CONTEXT}"

# Install Cilium
echo -e "${GREEN} Installing Cilium ${CILIUM_VERSION} in cilium namespace...${NC}"

# Get the real API server IP from the kubernetes endpoint so Cilium can reach it
# on worker nodes before the CNI (and therefore ClusterIP routing) is set up.
K8S_API_HOST=$(kubectl get endpoints kubernetes -o jsonpath='{.subsets[0].addresses[0].ip}')
K8S_API_PORT=$(kubectl get endpoints kubernetes -o jsonpath='{.subsets[0].ports[0].port}')
if [[ -z "$K8S_API_HOST" || -z "$K8S_API_PORT" ]]; then
  echo "ERROR: Kubernetes API endpoint is missing"
  exit 1
fi
echo "Using k8s API server: ${K8S_API_HOST}:${K8S_API_PORT}"

# Create cilium namespace
kubectl create namespace cilium --dry-run=client -o yaml | kubectl apply -f -

if helm status cilium -n cilium >/dev/null 2>&1; then
  cilium upgrade \
    --version "${CILIUM_VERSION}" \
    --namespace cilium \
    --helm-release-name cilium \
    --set ipam.mode=kubernetes \
    --set kubeProxyReplacement=false \
    --set enableHostFirewall=false \
    --set envoy.enabled=false \
    --set l7Proxy=false \
    --set k8sServiceHost="${K8S_API_HOST}" \
    --set k8sServicePort="${K8S_API_PORT}" \
    --wait
else
  cilium install \
    --version "${CILIUM_VERSION}" \
    --namespace cilium \
    --helm-release-name cilium \
    --set ipam.mode=kubernetes \
    --set kubeProxyReplacement=false \
    --set enableHostFirewall=false \
    --set envoy.enabled=false \
    --set l7Proxy=false \
    --set k8sServiceHost="${K8S_API_HOST}" \
    --set k8sServicePort="${K8S_API_PORT}" \
    --wait
fi

echo "Waiting for cilium namespace to exist..."
for i in {1..60}; do
  kubectl get namespace cilium >/dev/null 2>&1 && break
  sleep 2
done

kubectl get namespace cilium >/dev/null 2>&1 || {
  echo "ERROR: cilium namespace did not appear in time"
  exit 1
}

echo ""
echo "Waiting for Cilium to be ready..."
cilium status --wait --namespace cilium

echo -e "${GREEN}✓ Cilium installed successfully${NC}"
echo ""

# ---------- Flux ----------
echo -e "${GREEN}Installing Flux controllers in ${FLUX_NAMESPACE}...${NC}"

# Namespace
kubectl create namespace "${FLUX_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

# Install controllers (idempotent)
flux install --namespace "${FLUX_NAMESPACE}"

# Wait for core controllers to be ready
for controller in source-controller kustomize-controller helm-controller notification-controller; do
  kubectl -n "${FLUX_NAMESPACE}" rollout status "deploy/${controller}" --timeout=180s
done


# Summary
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Bootstrap Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Installed components:"
echo "  ✓ Cilium ${CILIUM_VERSION} (CNI + Network Policy) - namespace: cilium"
echo "  ✓ Flux controllers - namespace: ${FLUX_NAMESPACE}"
echo ""
echo "Useful commands:"
echo "  - Check Cilium status:   cilium status -n cilium"
echo "Hubble is configured later by the cluster Flux overlay."
