#!/bin/bash

set -euo pipefail

token="$(gopass show "odysseia/hellas/cloudflare" | head -n1 | tr -d '\r')"

kubectl create secret generic cloudflare-tunnel \
  --namespace=cloudflare-tunnel \
  --from-literal=tunnel-token="${token}" \
  --dry-run=client -o yaml > secret.yaml

sops -e -i secret.yaml
