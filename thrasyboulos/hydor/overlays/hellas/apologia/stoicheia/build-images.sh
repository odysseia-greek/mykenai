#!/usr/bin/env bash

set -euo pipefail

source_root="${1:-$(pwd)}"

build_image() {
  local component="$1"
  local tag="$2"
  local component_dir="${source_root}/${component}"

  if [[ ! -d "${component_dir}" ]]; then
    echo "error: component directory does not exist: ${component_dir}" >&2
    exit 1
  fi

  echo "Building ${component}:${tag}"
  (
    cd "${component_dir}"
    archimedes images single -t "${tag}" -m=false
  )
}

build_image alkibiades v0.1.5
build_image antisthenes v0.1.5
build_image aristippos v0.1.5
build_image aspasia v0.0.7
build_image kritias v0.1.5
build_image kriton v0.1.4
build_image xenofon v0.1.5
build_image sokrates v0.2.5
build_image meletos v0.1.3
build_image parmenides v0.2.5

echo "All Stoicheia images built successfully."
