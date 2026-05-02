#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")"

ZOT_USER="${ZOT_USER:-zot}"
ZOT_GOPASS_PATH="${ZOT_GOPASS_PATH:-odysseia/hellenistike/zot/password}"
ZOT_PASSWORD="$(gopass show -o "${ZOT_GOPASS_PATH}")"
ZOT_HTPASSWD="$(printf '%s\n' "${ZOT_PASSWORD}" | htpasswd -BinB "${ZOT_USER}")"

kubectl create secret generic zot-auth \
  --namespace=zot \
  --from-literal=username="${ZOT_USER}" \
  --from-literal=password="${ZOT_PASSWORD}" \
  --from-literal=htpasswd="${ZOT_HTPASSWD}" \
  --dry-run=client -o yaml > secret.yaml

sops -e -i secret.yaml
