#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")"

GITHUB_TOKEN_GOPASS_PATH="${GITHUB_TOKEN_GOPASS_PATH:-odysseia/github/tekton_token}"
GITHUB_TOKEN="$(gopass show -o "${GITHUB_TOKEN_GOPASS_PATH}")"

kubectl create secret generic github-status-token \
  --namespace=tekton-pipelines \
  --from-literal=token="${GITHUB_TOKEN}" \
  --dry-run=client -o yaml > github-status-token-secret.yaml

sops -e -i github-status-token-secret.yaml