# Minoan on Erebor

A minimal single-node k0s experiment on the physical Linux host `erebor`.
Shares the Cilium and Traefik bases with Romaioi, with one replica each,
L7 proxy, Envoy and Hubble (relay/UI) enabled, and Traefik exposed on NodePorts 30080/30443.
TopoLVM provisions local PVCs from a dedicated loop-backed VG. Rook-Ceph and OpenBao remain planned.

## Host setup

Review `lykourgos/ansible/inventories/minoan/hosts.yaml`. It uses Ansible's local connection: run it directly on Erebor. No SSH is needed.
The API binds to the host's primary IPv4 address; set `k0s_api_address` explicitly
if that is not the interface reachable from your workstation. Allow TCP 6443
from your workstation and the NodePorts from intended clients.

The playbook expects systemd Linux and sudo access. Host swap (including CachyOS
zram) is preserved: the default kubelet profile sets `failSwapOn: false` and
`memorySwap.swapBehavior: NoSwap`, so Kubernetes workloads do not use swap. Package
names handle Debian/Ubuntu and Arch-based hosts, including CachyOS. It runs `k0s sysinfo`
before installing the service. Erebor's actual OS, resources, existing services,
and spare disks have not been inspected. Use a host without an existing Kubernetes
installation and check that the pod/service CIDRs do not overlap your LAN/VPN.

From the repository root:

```sh
make -C lykourgos/ansible k0s-minoan ANSIBLE_FLAGS='-K'
kubectl --context=minoan-odysseia get nodes
```

The generated, gitignored kubeconfig is merged into `~/.kube/config` as
`minoan-odysseia` (admin user `minoan-odysseia-admin`). Existing entries and the
active context are preserved, and changes create a backup. Use
`kubectl config use-context minoan-odysseia` before the Helm/Flux commands below.
To merge an already generated config, run `make -C lykourgos/ansible merge-minoan-config ANSIBLE_FLAGS='-K'`.
The node will be NotReady until Cilium is installed. The k0s version is pinned
to `v1.36.4+k0s.1`; rerunning does not upgrade an installed binary.

## One-command bootstrap

After the Ansible host bootstrap, install the local CLIs (kubectl, Helm, Cilium,
and Flux), then commit and push the Minoan manifests to the branch Flux will use.
From the repository root:

```sh
make -C themistokles bootstrap-minoan BRANCH=main
```

This bootstraps Cilium 1.20.2 with Minoan values, installs Flux controllers, and
connects Flux to `themistokles/aer/minoan`, including Traefik, cert-manager and
TopoLVM. It targets `minoan-odysseia` explicitly without changing the active
context. Override with `KUBE_CONTEXT=...` if needed. This target exports the selected local private GPG key into `flux-system/sops-gpg`
before creating the Flux root Kustomization. Pass `SOPS_KEY_FP=your-fingerprint`.
Generate the Eleusinian test secret with the same key using:

```sh
make -C themistokles encrypt-minoan-eleusinian SOPS_KEY_FP=your-fingerprint
```

Commit and push the regenerated ciphertext and manifests before bootstrap. The
copied `telete.yaml` is still encrypted for its original recipient until regenerated.
If the manifests are on an experimental branch, pass `BRANCH=your-branch`.

## Networking and Flux

Install the shared base's Cilium chart version with the Minoan values first.
Flux will subsequently manage the same Helm release (`cilium` in `cilium`).
These commands require Helm, kubectl and the Flux CLI locally.

```sh
helm repo add cilium https://helm.cilium.io
helm repo update cilium
helm upgrade --install cilium cilium/cilium --version 1.20.2 \
  --namespace cilium --create-namespace \
  --set cluster.name=minoan --set operator.replicas=1 \
  --set ipam.mode=kubernetes --set kubeProxyReplacement=false \
  --values themistokles/ge/overlays/minoan/cilium/bootstrap-values.yaml --wait --timeout 10m
kubectl wait --for=condition=Ready nodes --all --timeout=300s
flux install
flux check
```

Commit and push the environment before creating the Git source. Use your branch
instead of `main` while experimenting if the files have not been merged yet.

```sh
flux create source git themistokles \
  --url=https://github.com/odysseia-greek/mykenai.git --branch=main --interval=1m
flux create kustomization themistokles \
  --source=themistokles --path=./themistokles/aer/minoan \
  --prune=true --interval=10m
flux get all
```

The Eleusinian Kustomization requires the matching SOPS GPG key. Point `traefik.minoan.odysseia-greek`
at Erebor to reach the dashboard at `http://traefik.minoan.odysseia-greek:30080/dashboard/`.

## Next experiments

- Rook-Ceph: choose an explicit unused device after inspecting `lsblk -f` on
  Erebor. The playbook installs `lvm2` but does not prepare or consume disks.
  Add an operator and single-node CephCluster/pool overlay with one monitor and
  replica size one for this disposable experiment. Configure CSI's kubelet root
  for k0s (`/var/lib/k0s/kubelet`). Do not enable automatic use of all devices.
- OpenBao: add a Minoan overlay after storage works, using a working StorageClass (TopoLVM initially, Ceph later).
  Review the existing `ge/base/delphi/openbao` as a starting point before wiring
  in Delphi applications and initialization/unseal. The Vault-based Delphi
  overlay and TopoLVM from Romaioi are intentionally not included.

References: [k0s installation](https://docs.k0sproject.io/latest/install/),
[Rook prerequisites](https://www.rook.io/docs/rook/latest/Getting-Started/Prerequisites/prerequisites/).

## Bounded disposable storage

Ansible preallocates `/var/lib/minoan/topolvm.img` (50 GiB by default) and creates
VG `minoan-pvc` exclusively on its loop device. Set `minoan_storage_gib` in the
inventory before first boot. Existing files are never implicitly resized, and
initial allocation preserves at least 10 GiB free on the host filesystem.
A systemd service restores the loop device and activates the VG before k0s at boot.
This is an experiment: host filesystem failure still affects the backing file.
PVC capacity is bounded; images, logs and `/var/lib/k0s` are not capped by this file.

Flux installs cert-manager and then TopoLVM, using `/var/lib/k0s/kubelet` and the
default `minoan-pvc-delete` StorageClass. One GiB of VG capacity is reserved.
Commit and push these manifests before asking Flux to reconcile them.

To erase the cluster and all PVCs, stop workloads and run this explicit reset.
The helper verifies the VG contains only the loop device backed by Minoan's file
before resetting k0s or removing storage. No physical disk is selected or wiped.
Do not manually delete an attached backing file.

```sh
make -C lykourgos/ansible reset-minoan ANSIBLE_FLAGS='-K -e minoan_reset_confirm=true'
# Reboot Erebor after k0s reset, then bootstrap again:
make -C lykourgos/ansible k0s-minoan ANSIBLE_FLAGS='-K'
```

For a future Ceph experiment, provision a separate unused device and migrate data;
the loop-backed TopoLVM VG is not automatically converted into Ceph storage.
