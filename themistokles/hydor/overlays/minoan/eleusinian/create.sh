#!/bin/bash
set -euo pipefail
for tool in kubectl sops; do
  command -v "$tool" >/dev/null || { echo "Required command missing: $tool" >&2; exit 1; }
done
: "${SOPS_KEY_FP:?Set SOPS_KEY_FP to the intended Minoan GPG fingerprint}"
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
output=$(mktemp "$script_dir/.telete.XXXXXX.yaml")
trap 'rm -f "$output"' EXIT
chmod 600 "$output"
kubectl create secret generic telete \
  --namespace=eleusinian \
  --from-literal=hierophant='flux' \
  --from-literal=mystery='revealed only within the telesterion' \
  --dry-run=client -o yaml | \
  sops --encrypt --pgp "$SOPS_KEY_FP" --encrypted-regex '^(data|stringData)$' \
    --input-type yaml --output-type yaml /dev/stdin > "$output"
mv -- "$output" "$script_dir/telete.yaml"
echo "Encrypted Eleusinian secret using $SOPS_KEY_FP"
