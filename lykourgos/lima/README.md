# k0s on Lima

Local k0s clusters running in Lima VMs as an alternative to k3d. Uses Debian 13 Trixie (similar to Raspbian) and provides static IPs in the 192.168.x.x range.

Each VM includes a 30GB additional disk (`pyxis`) for TopoLVM, matching your ACC and production setups.

## Prerequisites

```bash
# Install Lima
brew install lima

# Verify installation
limactl --version
```

## Quick Start

### Single-Node Cluster (with Makefile)

```bash
cd lykourgos/lima

# Start single-node cluster (automatically creates 30GB disk)
make single-start

# Get kubeconfig
make kubeconfig-single
export KUBECONFIG=~/.kube/k0s-byzantium-config

# Update /etc/hosts
make update-hosts

# Verify cluster
kubectl get nodes
```

### HA Cluster (with Makefile)

```bash
cd lykourgos/lima

# Start HA cluster (automatically creates 3x 30GB disks)
make ha-start

# Get kubeconfig
make kubeconfig-ha
export KUBECONFIG=~/.kube/k0s-byzantium-ha-config

# Update /etc/hosts
make update-hosts

# Verify cluster
kubectl get nodes -o wide
```

## Using Ansible

### Single-Node with Ansible

```bash
cd lykourgos/ansible
ansible-playbook -i inventories/lima/hosts.yaml playbooks/k0s-lima-single.yaml
```

### HA Cluster with Ansible

```bash
cd lykourgos/ansible
ansible-playbook -i inventories/lima/hosts.yaml playbooks/k0s-lima-ha.yaml
```

## IP Addresses

| Cluster Type | Node | IP | Hostname |
|--------------|------|-----|----------|
| Single | k0s-byzantium | 192.168.105.2 | k0s-byzantium.odysseia-greek |
| HA | k0s-controller | 192.168.105.10 | k0s-byzantium.odysseia-greek |
| HA | k0s-worker1 | 192.168.105.11 | - |
| HA | k0s-worker2 | 192.168.105.12 | - |

## /etc/hosts Configuration

### Single-Node
```
192.168.105.2 k0s-byzantium.odysseia-greek byzantium.odysseia-greek
```

### HA Cluster
```
192.168.105.10 k0s-byzantium.odysseia-greek byzantium.odysseia-greek
```

## Resource Allocation

### Single-Node
- CPU: 4 cores
- Memory: 8 GiB
- Root Disk: 50 GiB
- Additional Disk (pyxis): 30 GiB (for TopoLVM)

### HA Controller
- CPU: 4 cores
- Memory: 8 GiB
- Root Disk: 50 GiB
- Additional Disk (pyxis-controller): 30 GiB (for TopoLVM)

### HA Workers
- CPU: 3 cores each
- Memory: 6 GiB each
- Root Disk: 40 GiB each
- Additional Disk (pyxis-worker1/2): 30 GiB each (for TopoLVM)

**Total HA**: 10 cores, 20 GiB RAM, 130 GiB root + 90 GiB additional

## Makefile Commands

```bash
# Single-node
make single-start        # Create disk & start cluster
make single-stop         # Stop cluster
make single-delete       # Delete cluster (keeps disk)
make kubeconfig-single   # Get kubeconfig

# HA cluster
make ha-start            # Create disks & start cluster
make ha-stop             # Stop cluster
make ha-delete           # Delete cluster (keeps disks)
make kubeconfig-ha       # Get kubeconfig

# Disk management
make list-disks          # List all pyxis disks
make delete-disks        # Delete all pyxis disks
make create-disk-single  # Manually create single disk
make create-disk-ha      # Manually create HA disks

# General
make status              # Show VM status
make clean               # Delete VMs and disks
make update-hosts        # Update /etc/hosts

# Shell & logs
make shell-single        # Shell into single-node
make shell-controller    # Shell into controller
make logs-single         # Follow controller logs
make logs-worker1        # Follow worker1 logs
```

## Common Commands

```bash
# List VMs & disks
limactl list
limactl disk ls

# Disk operations (done automatically by Makefile)
limactl disk create pyxis --size 30G
limactl disk delete pyxis

# Shell into VM
limactl shell k0s-byzantium

# View logs
limactl shell k0s-byzantium sudo journalctl -u k0scontroller -f

# Get cluster status
limactl shell k0s-byzantium sudo k0s kubectl get nodes -o wide
limactl shell k0s-byzantium sudo k0s status

# Check disk in VM
limactl shell k0s-byzantium lsblk
limactl shell k0s-byzantium sudo fdisk -l
```

## Installing CNI (Cilium)

After cluster is up, install Cilium:

```bash
# Install Cilium CLI
brew install cilium-cli

# Install Cilium to cluster
cilium install --version 1.15.0

# Verify
cilium status
kubectl get pods -n kube-system
```

## Cleanup

```bash
# With Makefile (deletes VMs and disks)
make clean

# Or manually
# Single-node
make single-delete   # Keeps disk
make delete-disks    # Delete disk

# HA cluster
make ha-delete       # Keeps disks
make delete-disks    # Delete disks

# Remove from /etc/hosts
sudo sed -i.bak '/k0s-byzantium.odysseia-greek/d' /etc/hosts
```

## Installing TopoLVM

The additional disks are ready for TopoLVM setup, matching your ACC/production environment:

```bash
# Check disk is available
limactl shell k0s-byzantium lsblk

# You should see a disk like /dev/vdb (30GB)
# Set up LVM similar to your Raspberry Pi setup
limactl shell k0s-byzantium sudo pvcreate /dev/vdb
limactl shell k0s-byzantium sudo vgcreate topolvm /dev/vdb

# Install TopoLVM via Helm or kubectl
# (Use your existing TopoLVM configurations)
```

## Comparison with k3d

| Feature | k3d | Lima k0s |
|---------|-----|----------|
| Container Runtime | Docker | Lima (QEMU/VZ) |
| IP Range | 127.0.0.1 | 192.168.105.x |
| OS | Alpine (container) | Debian 13 (VM) |
| Matches Raspberry Pi | No | Yes (Debian ≈ Raspbian) |
| Resource Isolation | Container | Full VM |
| Network Access | localhost | Network IP |
| Startup Time | Fast (~30s) | Slower (~2-3min) |
| Resource Overhead | Low | Higher (VM overhead) |
| Additional Disks | No | Yes (30GB for TopoLVM) |
| LVM Support | No | Yes |

## Troubleshooting

### VM won't start
```bash
# Check Lima logs
limactl shell k0s-byzantium cat /var/log/cloud-init-output.log

# Check system logs
limactl shell k0s-byzantium sudo journalctl -xe
```

### Network issues
```bash
# Check IP address
limactl shell k0s-byzantium ip addr show

# Test connectivity
ping 192.168.105.2
```

### k0s not starting
```bash
# Check k0s status
limactl shell k0s-byzantium sudo systemctl status k0scontroller

# View logs
limactl shell k0s-byzantium sudo journalctl -u k0scontroller -n 100

# Restart k0s
limactl shell k0s-byzantium sudo systemctl restart k0scontroller
```

### Worker won't join
```bash
# Verify token
limactl shell k0s-controller cat /tmp/worker-token.txt

# Check controller connectivity from worker
limactl shell k0s-worker1 ping -c 3 192.168.105.10
limactl shell k0s-worker1 cat /etc/hosts

# Generate new token (expires in 24h by default)
limactl shell k0s-controller sudo k0s token create --role=worker --expiry=24h
```
