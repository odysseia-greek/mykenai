#!/bin/bash

set -euo pipefail

token="$(gopass show "odysseia/hellas/cloudflare tunnel")"

kubectl create secret generic cloudflare-tunnel \
  --namespace=cloudflare-tunnel \
  --from-literal=tunnel-token="${token}" \
  --dry-run=client -o yaml > secret.yaml

sops -e -i secret.yaml
