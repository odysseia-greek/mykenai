#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")"

ZOT_REGISTRY="${ZOT_REGISTRY:-zot.zot.svc.cluster.local:5000}"
ZOT_REGISTRIES="${ZOT_REGISTRIES:-${ZOT_REGISTRY}}"
ZOT_USER="${ZOT_USER:-zot}"
ZOT_GOPASS_PATH="${ZOT_GOPASS_PATH:-odysseia/hellenistike/zot/password}"
ZOT_PASSWORD="$(gopass show -o "${ZOT_GOPASS_PATH}")"
ZOT_AUTH="$(printf '%s:%s' "${ZOT_USER}" "${ZOT_PASSWORD}" | base64)"
DOCKER_CONFIG="$(
  jq -n \
    --arg registries "${ZOT_REGISTRIES}" \
    --arg username "${ZOT_USER}" \
    --arg password "${ZOT_PASSWORD}" \
    --arg auth "${ZOT_AUTH}" \
    '{
      auths: (
        $registries
        | split(",")
        | map(gsub("^\\s+|\\s+$"; ""))
        | map(select(length > 0))
        | map({key: ., value: {username: $username, password: $password, auth: $auth}})
        | from_entries
      )
    }'
)"

kubectl create secret generic registry-creds \
  --namespace=tekton-pipelines \
  --type=kubernetes.io/dockerconfigjson \
  --from-literal=.dockerconfigjson="${DOCKER_CONFIG}" \
  --dry-run=client -o yaml > secret.yaml

sops -e -i secret.yaml
