# Storage Experiment: Rook-Ceph → Longhorn Migration

## Context

This document captures the attempt to deploy distributed storage on the management cluster **hellenistike**, and the eventual decision to migrate from Rook-Ceph to Longhorn.

---

## Cluster Overview

### Nodes (Kubernetes manifests)

    apiVersion: v1
    kind: Node
    metadata:
      name: pella-hellenistike
      annotations:
        kustomize.toolkit.fluxcd.io/prune: disabled
        kustomize.toolkit.fluxcd.io/ssa: disabled
      labels:
        cilium.io/envoy-capable: "false"
    ---
    apiVersion: v1
    kind: Node
    metadata:
      name: antioch-hellenistike
      annotations:
        kustomize.toolkit.fluxcd.io/prune: disabled
        kustomize.toolkit.fluxcd.io/ssa: disabled
      labels:
        cilium.io/envoy-capable: "false"
    ---
    apiVersion: v1
    kind: Node
    metadata:
      name: alexandreia-hellenistike
      annotations:
        kustomize.toolkit.fluxcd.io/prune: disabled
        kustomize.toolkit.fluxcd.io/ssa: disabled
      labels:
        cilium.io/envoy-capable: "false"

### Hardware

| Node                     | IP              | RAM  | Storage   |
|--------------------------|-----------------|------|-----------|
| pella-hellenistike       | 192.168.1.131   | 8GB  | 256GB SSD |
| alexandreia-hellenistike | 192.168.1.132   | 8GB  | 256GB SSD |
| antioch-hellenistike     | 192.168.1.133   | 4GB  | 256GB SSD |

---

## Initial Goal

Deploy Rook-Ceph to provide:

- Replicated block storage
- Distributed storage behavior
- A platform to learn Ceph internals

---

## Observations During Rook-Ceph Deployment

### Resource Usage

Memory usage (approximate):

- pella: 32%
- alexandreia: 37%
- antioch: 82%

Estimated Ceph memory footprint per node:
- ~1.5GB → 3.1GB

CPU load averages:

- pella: ~1.3
- alexandreia: ~1.5
- antioch: ~3.3

---

### Key Issues

1. Uneven resource pressure
    - `antioch-hellenistike` (4GB RAM) became the bottleneck
    - High memory pressure from OSD + Ceph daemons

2. Operational instability
    - Difficulty keeping HelmRelease in sync
    - Sensitive to resource fluctuations

3. Cluster overhead too high
    - Even with minimal configuration
    - Significant baseline resource usage

---

## Conclusion on Rook-Ceph

While functional, the setup proved:

- Too resource intensive for a mixed-node Raspberry Pi cluster
- Operationally heavy for the intended use case

Better suited for:

- Uniform nodes
- ≥8–16GB RAM per node
- Dedicated storage clusters

---

## Migration to Longhorn

Switched to Longhorn with:

- Default configuration
- Minor sensible adjustments

### Result

- Deployment succeeded on first attempt
- No significant tuning required
- Immediate operational stability

---

## Post-Migration Observations

(Cluster currently powered down; based on observed behavior)

- Significantly lower resource usage
- No noticeable CPU pressure spikes
- No memory saturation on lower-spec node
- Predictable and stable operation

---

## Final Assessment

### Rook-Ceph

**Pros**
- True distributed storage system
- Advanced features (self-healing, data distribution)

**Cons**
- High resource requirements
- Complex operational model
- Poor fit for heterogeneous low-resource clusters

---

### Longhorn

**Pros**
- Lightweight
- Kubernetes-native
- Easy to deploy and operate
- Fits homelab constraints

**Cons**
- Simpler replication model
- Less advanced data distribution

---

## Decision

Adopt Longhorn as the default storage solution for the `hellenistike` cluster.

Ceph remains a future option for:

- Dedicated hardware
- Higher-memory nodes
- Learning and experimentation

---

## Key Insight

Not all correct solutions are appropriate solutions.

Ceph was technically valid, but Longhorn aligned better with:

- hardware constraints
- operational simplicity
- overall system goals

---

## Future Work

- Revisit Ceph on:
    - dedicated hardware
    - or VM-based lab environment

- Validate Longhorn failure scenarios:
    - disk removal
    - node loss
    - replica rebuild behavior  