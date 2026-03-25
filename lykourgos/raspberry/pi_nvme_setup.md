# Raspberry Pi NVMe Provisioning Guide (Headless + LVM + SSH)

## Overview

This guide walks through:

- Flashing Raspberry Pi OS Lite to NVMe
- Creating a split disk layout:
  - OS partition
  - LVM storage pool (`pyxis`)
- Enabling headless SSH with key authentication
- Booting from NVMe
- Verifying and troubleshooting

---

## Final Expected State

### Disk Layout

```
nvme0n1
├─nvme0n1p1   512M   /boot/firmware
├─nvme0n1p2   ~100G  /
└─nvme0n1p3   ~rest  LVM (pyxis)
```

### LVM

```
sudo vgs

VG    #PV #LV #SN Attr   VSize    VFree
pyxis   1   0   0 wz--n- <145.34g <145.34g
```

---

## Step 1 — Flash OS to NVMe

```
wget https://downloads.raspberrypi.com/raspios_lite_arm64_latest -O raspios.img.xz
unxz raspios.img.xz

sudo dd if=raspios.img of=/dev/nvme0n1 bs=4M status=progress conv=fsync
```

---

## Step 2 — Partition Disk

```
sudo parted /dev/nvme0n1
```

Inside:

```
resizepart 2 100GB
mkpart primary 100GB 100%
```

---

## Step 3 — Resize Filesystem

```
sudo e2fsck -f /dev/nvme0n1p2
sudo resize2fs /dev/nvme0n1p2
```

---

## Step 4 — Create LVM (pyxis)

```
sudo apt install lvm2 -y

sudo pvcreate /dev/nvme0n1p3
sudo vgcreate pyxis /dev/nvme0n1p3
```

---

## Step 5 — Prepare Headless Boot

Mount partitions from SD system:

```
sudo mount /dev/nvme0n1p2 /mnt/nvme-root
sudo mount /dev/nvme0n1p1 /mnt/nvme-boot
```

### Enable SSH

```
sudo touch /mnt/nvme-boot/ssh
```

### Create user

```
echo "pi:$(openssl passwd -6 raspberry)" | sudo tee /mnt/nvme-boot/userconf.txt
```

### Add SSH key

```
sudo mkdir -p /mnt/nvme-root/home/pi/.ssh
sudo nano /mnt/nvme-root/home/pi/.ssh/authorized_keys
```

Paste your public key.

### Fix permissions

```
sudo chmod 700 /mnt/nvme-root/home/pi/.ssh
sudo chmod 600 /mnt/nvme-root/home/pi/.ssh/authorized_keys
sudo chown -R 1000:1000 /mnt/nvme-root/home/pi/.ssh
```

---

## Step 6 — Boot from NVMe

- Power off
- Remove SD card
- Boot

SSH:

```
ssh pi@raspberrypi.local
```

---

## Step 7 — Set Hostname

```
sudo hostnamectl set-hostname athenai-hellas
sudo nano /etc/hosts
```

Change:

```
127.0.1.1   athenai-hellas
```

---

## Step 8 — Secure SSH

```
sudo nano /etc/ssh/sshd_config
```

Set:

```
PasswordAuthentication no
PubkeyAuthentication yes
```

Restart:

```
sudo systemctl restart ssh
```

---

## Verification

### Disk

```
lsblk
```

Expected:

```
nvme0n1p2 mounted on /
```

### Filesystem

```
lsblk -f
```

Expected:

```
nvme0n1p3 LVM2_member
```

### LVM

```
sudo vgs
```

Expected:

```
pyxis
```

---

## Troubleshooting

### SSH asks for password

- Check authorized_keys exists
- Check permissions (700 / 600)
- Check ownership (pi:pi)

---

### Cannot SSH

Check network:

```
ping raspberrypi.local
```

Fallback:

```
arp -a
```

---

### LVM not visible

```
sudo vgscan
sudo vgchange -ay
```

---

### Wrong disk wiped

Always verify:

```
lsblk
```

- SD = mmcblk0
- NVMe = nvme0n1

---

## Notes

- Do NOT format `nvme0n1p3`
- TopoLVM will use raw LVM space
- This setup mirrors production storage separation

---

## Outcome

You now have:

- NVMe boot node
- SSH key-only access
- Dedicated storage pool (`pyxis`)
- Reproducible base for cluster nodes
