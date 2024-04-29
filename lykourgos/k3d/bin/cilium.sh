#!/bin/sh

set -e

echo "Mounting bpf on node"
mount bpffs -t bpf /sys/fs/bpf
mount --make-shared /sys/fs/bpf

echo "Mounting cgroups v2 to /run/cilium/cgroupv2 on node"
mkdir -p /run/cilium/cgroupv2
mount -t cgroup2 none /run/cilium/cgroupv2
mount --make-shared /run/cilium/cgroupv2/

echo "Mounted needed directories for Cilium. You can now install cilium using the cli or helm"
