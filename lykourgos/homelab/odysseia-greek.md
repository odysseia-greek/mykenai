# HOMELAB

## Overview

This document describes the current state and evolution of the homelab Kubernetes environments.

Three primary clusters exist:

- **hellenistike** → management cluster (Raspberry Pi)
- **hellas-odysseia** → homelab workload cluster (physical nodes)
- **romaioi** → development cluster (Lima VMs on macOS)

The current topology is also available as a [D2 diagram](./odysseia-greek.d2).

---

## Original Plan (Outdated)

Initial intent:

- Bootstrap Raspberry Pi cluster using Ansible
- Deploy:
  - rook-ceph
  - cilium
  - traefik
  - etcd

Cluster layout:

### Servers

| Node            | Hardware |
|-----------------|----------|
| k3s-s-athenai   | rpi5 8GB |
| k3s-s-sparta    | rpi4 8GB |
| k3s-s-syrakousai| rpi5 8GB |

### Workers

| Node            | Hardware |
|-----------------|----------|
| k3s-w-thebai    | rpi4 8GB |
| k3s-w-korinth   | rpi5 4GB |
| k3s-w-argos     | rpi4 4GB |
| k3s-w-taras     | rpi4 4GB |

> ⚠️ This layout is no longer in sync with reality.

---

## Current State

### Cluster: hellenistike (Management)

#### Nodes

| Node                     | IP              | RAM  | Storage   |
|--------------------------|-----------------|------|-----------|
| pella-hellenistike       | 192.168.1.131   | 8GB  | 256GB SSD |
| alexandreia-hellenistike | 192.168.1.132   | 8GB  | 256GB SSD |
| antioch-hellenistike     | 192.168.1.133   | 4GB  | 256GB SSD |

#### Characteristics

- Mixed hardware (8GB + 4GB node)
- SSD-backed storage on all nodes
- Acts as **management cluster**

---

### Hellenistike Deployment Structure

Path: `mykenai/themistokles/ge/overlays/hellenistike`

Deployed components:

- cert-manager
- cilium
- eleusinian (SOPS validation namespace)
- kaniko
- labels
- longhorn (replaced rook-ceph)
- tekton
- traefik
- zot

---

### Cluster: hellas-odysseia

#### Nodes

| Node             | Status | Role          | Age | Kubernetes version | NVMe | TopoLVM |
|------------------|--------|---------------|-----|--------------------|------|---------|
| athenai-hellas   | Ready  | worker        | 68d | v1.35.2+k0s        | Yes  | Yes     |
| korinthos-hellas | Ready  | worker        | 25d | v1.35.2+k0s        | No   | No      |
| sparta-hellas    | Ready  | control-plane | 68d | v1.35.2+k0s        | Yes  | Yes     |
| thebai-hellas    | Ready  | worker        | 68d | v1.35.2+k0s        | Yes  | Yes     |

#### Characteristics

- Four-node physical k0s cluster
- `sparta-hellas` provides the control plane
- All nodes except `korinthos-hellas` have NVMe storage and run TopoLVM
- `korinthos-hellas` participates as a compute worker without TopoLVM-backed local storage

#### Path

Path: `mykenai/themistokles/ge/overlays/hellas`

---

## Storage Decision

### Previous Attempt

- Tried deploying Rook-Ceph
- Observed:
  - High memory usage (up to ~3GB per node)
  - Heavy CPU load
  - Instability on 4GB node (antioch)
  - Difficult HelmRelease reconciliation

### Outcome

- Determined cluster is not suitable for Ceph
- Switched to Longhorn

### Result

- Stable deployment on first attempt
- Significantly lower resource usage
- Better operational fit for hardware

---

## Cluster: romaioi (Development)

#### Description

- Three-node development cluster
- Runs on Lima VMs on macOS
- Uses k0s
- Uses TopoLVM for local storage

#### Nodes

| Node           | Status | Role          | Age  | Kubernetes version | Runtime |
|----------------|--------|---------------|------|--------------------|---------|
| lima-byzantion | Ready  | control-plane | 154m | v1.35.2+k0s        | Lima VM |
| lima-nikaia    | Ready  | worker        | 153m | v1.35.2+k0s        | Lima VM |
| lima-trapezous | Ready  | worker        | 153m | v1.35.2+k0s        | Lima VM |

#### Path

Path: `mykenai/themistokles/ge/overlays/romaioi`

#### Purpose

- Local development
- Fast iteration environment
- Multi-node and TopoLVM testing before changes reach the physical clusters

---

## Key Insight

The homelab evolved from:

> “replicate production-grade distributed systems”

to:

> “build a stable, understandable platform to iterate on”

---

## Future Direction

- Keep Longhorn as default storage for Pi clusters
- Revisit Ceph only when:
  - uniform hardware is available
  - ≥16GB RAM per node
  - or via VM-based lab setup

- Continue refining:
  - Flux structure
  - Kustomize overlays
  - cluster separation (dev vs mgmt)
  - TopoLVM operations across physical and Lima-based clusters

---

## Notes

- `eleusinian` namespace is used for SOPS validation
- `_lethe/` holds deprecated or unused configurations
- Naming follows Ancient Greek theme for consistency
