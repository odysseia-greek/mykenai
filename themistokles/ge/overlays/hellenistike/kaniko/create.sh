#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")"

ZOT_REGISTRY="${ZOT_REGISTRY:-registry.hellenistike.odysseia-greek:30080}"
ZOT_USER="${ZOT_USER:-zot}"
ZOT_GOPASS_PATH="${ZOT_GOPASS_PATH:-odysseia/hellenistike/zot/password}"
ZOT_PASSWORD="$(gopass show -o "${ZOT_GOPASS_PATH}")"
ZOT_AUTH="$(printf '%s:%s' "${ZOT_USER}" "${ZOT_PASSWORD}" | base64)"

kubectl create secret generic registry-creds \
  --namespace=tekton-pipelines \
  --type=kubernetes.io/dockerconfigjson \
  --from-literal=.dockerconfigjson="{\"auths\":{\"${ZOT_REGISTRY}\":{\"username\":\"${ZOT_USER}\",\"password\":\"${ZOT_PASSWORD}\",\"auth\":\"${ZOT_AUTH}\"}}}" \
  --dry-run=client -o yaml > secret.yaml

sops -e -i secret.yaml
