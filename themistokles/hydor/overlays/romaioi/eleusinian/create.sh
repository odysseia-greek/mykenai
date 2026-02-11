#!/bin/bash

kubectl create secret generic telete \
  --namespace=eleusinian \
  --from-literal=hierophant='flux' \
  --from-literal=mystery='revealed only within the telesterion' \
  --dry-run=client -o yaml > telete.yaml

sops -e -i telete.yaml