# HOMELAB

## Overview

This document describes the current state and evolution of the homelab Kubernetes environments.

Two primary clusters exist:

- **hellenistike** → management cluster (Raspberry Pi)
- **romaioi** → development cluster (Lima VM on macOS)

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

### Deployment Structure

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

- Single-node development cluster
- Runs on Lima VM on macOS
- Uses k0s

#### Path

Path: `mykenai/themistokles/ge/overlays/romaioi`

#### Specs

| Resource | Value  |
|----------|--------|
| CPU      | 12     |
| Memory   | 12 GiB |
| Disk     | 25 GiB |

#### Purpose

- Local development
- Fast iteration environment

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

---

## Notes

- `eleusinian` namespace is used for SOPS validation
- `_lethe/` holds deprecated or unused configurations
- Naming follows Ancient Greek theme for consistency
