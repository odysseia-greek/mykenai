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

build_image alexandros v0.3.3
build_image antigonos v0.0.4
build_image dareios v0.0.2
build_image demokritos v0.2.5
build_image eukleides v0.0.5
build_image hefaistion v0.0.4
build_image parmenion v0.0.4
build_image perdikkas v0.0.5
build_image ptolemaios v0.0.4

echo "All Makedonia images built successfully."
