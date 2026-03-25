#!/usr/bin/env bash
set -Eeuo pipefail

# Raspberry Pi NVMe provisioning script
#
# What it does:
# - flashes a Raspberry Pi OS image to a target disk
# - expands root partition to a chosen size
# - creates a third partition for LVM
# - creates VG "pyxis"
# - injects headless boot config:
#   - SSH enabled
#   - hostname set
#   - user created
#   - SSH public key installed
#
# Run this from a Linux system that can see the target disk, for example:
# - a Raspberry Pi booted from SD
# - another Linux host
#
# Example:
# sudo ./provision_pi_nvme.sh \
#   --device /dev/nvme0n1 \
#   --image /path/to/raspios.img \
#   --hostname athenai-hellas \
#   --root-size 100G \
#   --ssh-pubkey ~/.ssh/id_ed25519.pub
#
# Notes:
# - This script assumes Raspberry Pi OS Lite image layout:
#   p1 = boot, p2 = root
# - It creates p3 as an LVM PV and VG named "pyxis"
# - It enables SSH by placing /boot/firmware/ssh
# - It creates userconf.txt with a temporary password hash
# - After first boot, disable password auth in sshd_config if desired

usage() {
  cat <<'EOF'
Usage:
  sudo ./provision_pi_nvme.sh \
    --device /dev/nvme0n1 \
    --image /path/to/raspios.img \
    --hostname athenai-hellas \
    --root-size 100G \
    --ssh-pubkey /path/to/id_ed25519.pub \
    [--username pi] \
    [--password-hash '$6$...'] \
    [--vg-name pyxis] \
    [--yes]

Required:
  --device         target disk, e.g. /dev/nvme0n1
  --image          uncompressed Raspberry Pi OS image (.img)
  --hostname       hostname to configure
  --root-size      size for root partition, e.g. 80G, 100G
  --ssh-pubkey     path to public key file

Optional:
  --username       Linux user to create; default: pi
  --password-hash  precomputed SHA-512 password hash for userconf.txt
                   if omitted, a temporary bootstrap password "raspberry" is hashed
  --vg-name        LVM volume group name; default: pyxis
  --yes            do not prompt for confirmation

Examples:
  openssl passwd -6 raspberry
  ssh-keygen -t ed25519

EOF
}

DEVICE=""
IMAGE=""
HOSTNAME_SET=""
ROOT_SIZE=""
SSH_PUBKEY_FILE=""
USERNAME="pi"
PASSWORD_HASH=""
VG_NAME="pyxis"
ASSUME_YES="false"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --device) DEVICE="${2:-}"; shift 2 ;;
    --image) IMAGE="${2:-}"; shift 2 ;;
    --hostname) HOSTNAME_SET="${2:-}"; shift 2 ;;
    --root-size) ROOT_SIZE="${2:-}"; shift 2 ;;
    --ssh-pubkey) SSH_PUBKEY_FILE="${2:-}"; shift 2 ;;
    --username) USERNAME="${2:-}"; shift 2 ;;
    --password-hash) PASSWORD_HASH="${2:-}"; shift 2 ;;
    --vg-name) VG_NAME="${2:-}"; shift 2 ;;
    --yes) ASSUME_YES="true"; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage; exit 1 ;;
  esac
done

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Missing required command: $1" >&2
    exit 1
  }
}

cleanup() {
  set +e
  sync
  if mountpoint -q /mnt/pi-root; then umount /mnt/pi-root; fi
  if mountpoint -q /mnt/pi-boot; then umount /mnt/pi-boot; fi
}
trap cleanup EXIT

fail() {
  echo "ERROR: $*" >&2
  exit 1
}

[[ $EUID -eq 0 ]] || fail "Run as root or with sudo."
[[ -n "$DEVICE" ]] || fail "--device is required"
[[ -n "$IMAGE" ]] || fail "--image is required"
[[ -n "$HOSTNAME_SET" ]] || fail "--hostname is required"
[[ -n "$ROOT_SIZE" ]] || fail "--root-size is required"
[[ -n "$SSH_PUBKEY_FILE" ]] || fail "--ssh-pubkey is required"

[[ -b "$DEVICE" ]] || fail "Device does not exist or is not a block device: $DEVICE"
[[ -f "$IMAGE" ]] || fail "Image file not found: $IMAGE"
[[ -f "$SSH_PUBKEY_FILE" ]] || fail "SSH public key file not found: $SSH_PUBKEY_FILE"

for cmd in dd parted e2fsck resize2fs lsblk mount umount sync wipefs sed awk grep tee chmod chown mkdir touch; do
  require_cmd "$cmd"
done

if ! command -v pvcreate >/dev/null 2>&1 || ! command -v vgcreate >/dev/null 2>&1; then
  echo "LVM tools not found. Install lvm2 first." >&2
  echo "Example: sudo apt update && sudo apt install -y lvm2" >&2
  exit 1
fi

SSH_PUBKEY="$(<"$SSH_PUBKEY_FILE")"
[[ "$SSH_PUBKEY" == ssh-* ]] || fail "Public key does not look like a valid OpenSSH public key."

if [[ -z "$PASSWORD_HASH" ]]; then
  require_cmd openssl
  PASSWORD_HASH="$(openssl passwd -6 raspberry)"
fi

echo "=== Provisioning plan ==="
echo "Device     : $DEVICE"
echo "Image      : $IMAGE"
echo "Hostname   : $HOSTNAME_SET"
echo "Root size  : $ROOT_SIZE"
echo "User       : $USERNAME"
echo "VG name    : $VG_NAME"
echo "SSH pubkey : $SSH_PUBKEY_FILE"
echo
lsblk "$DEVICE" || true
echo

if [[ "$ASSUME_YES" != "true" ]]; then
  read -r -p "This will destroy all data on $DEVICE. Continue? [yes/NO] " reply
  [[ "$reply" == "yes" ]] || fail "Aborted."
fi

echo "==> Unmounting any mounted partitions on target"
while read -r part _; do
  [[ "$part" == NAME ]] && continue
  umount "/dev/$part" 2>/dev/null || true
done < <(lsblk -ln -o NAME,MOUNTPOINT "$DEVICE")

echo "==> Wiping old signatures"
wipefs -a "$DEVICE"

echo "==> Flashing image to disk"
dd if="$IMAGE" of="$DEVICE" bs=4M status=progress conv=fsync
sync
partprobe "$DEVICE" || true
sleep 2

BOOT_PART="${DEVICE}p1"
ROOT_PART="${DEVICE}p2"

if [[ "$DEVICE" =~ mmcblk|nvme ]]; then
  BOOT_PART="${DEVICE}p1"
  ROOT_PART="${DEVICE}p2"
  DATA_PART="${DEVICE}p3"
else
  BOOT_PART="${DEVICE}1"
  ROOT_PART="${DEVICE}2"
  DATA_PART="${DEVICE}3"
fi

[[ -b "$BOOT_PART" ]] || fail "Boot partition not found after flashing: $BOOT_PART"
[[ -b "$ROOT_PART" ]] || fail "Root partition not found after flashing: $ROOT_PART"

echo "==> Expanding root partition to $ROOT_SIZE"
parted -s "$DEVICE" unit GiB print || true
parted -s "$DEVICE" resizepart 2 "$ROOT_SIZE"
partprobe "$DEVICE" || true
sleep 2

echo "==> Checking and resizing root filesystem"
e2fsck -f -y "$ROOT_PART"
resize2fs "$ROOT_PART"

echo "==> Creating LVM partition from $ROOT_SIZE to 100%"
parted -s "$DEVICE" mkpart primary "$ROOT_SIZE" 100%
partprobe "$DEVICE" || true
sleep 2

[[ -b "$DATA_PART" ]] || fail "Data partition not found after creation: $DATA_PART"

echo "==> Creating LVM PV and VG"
pvcreate "$DATA_PART"
vgcreate "$VG_NAME" "$DATA_PART"

echo "==> Mounting root and boot partitions"
mkdir -p /mnt/pi-root /mnt/pi-boot
mount "$ROOT_PART" /mnt/pi-root
mount "$BOOT_PART" /mnt/pi-boot

echo "==> Enabling SSH on first boot"
touch /mnt/pi-boot/ssh

echo "==> Creating bootstrap user config"
printf '%s:%s\n' "$USERNAME" "$PASSWORD_HASH" > /mnt/pi-boot/userconf.txt

echo "==> Setting hostname"
printf '%s\n' "$HOSTNAME_SET" > /mnt/pi-root/etc/hostname

cat > /mnt/pi-root/etc/hosts <<EOF
127.0.0.1   localhost
127.0.1.1   $HOSTNAME_SET

::1         localhost ip6-localhost ip6-loopback
ff02::1     ip6-allnodes
ff02::2     ip6-allrouters
EOF

echo "==> Installing SSH public key for user $USERNAME"
USER_HOME="/mnt/pi-root/home/$USERNAME"
mkdir -p "$USER_HOME/.ssh"
printf '%s\n' "$SSH_PUBKEY" > "$USER_HOME/.ssh/authorized_keys"
chmod 700 "$USER_HOME/.ssh"
chmod 600 "$USER_HOME/.ssh/authorized_keys"

USER_UID="$(chroot /mnt/pi-root id -u "$USERNAME" 2>/dev/null || true)"
USER_GID="$(chroot /mnt/pi-root id -g "$USERNAME" 2>/dev/null || true)"

if [[ -z "$USER_UID" || -z "$USER_GID" ]]; then
  # Raspberry Pi OS default first user is usually UID/GID 1000.
  USER_UID="1000"
  USER_GID="1000"
fi

chown -R "$USER_UID:$USER_GID" "$USER_HOME/.ssh"

echo "==> Syncing changes"
sync

echo
echo "=== Provisioning completed ==="
echo
echo "Expected layout:"
lsblk -f "$DEVICE"
echo
echo "Expected LVM:"
pvs
vgs
echo
echo "Next steps:"
echo "1. Power off the machine."
echo "2. Remove the SD card."
echo "3. Boot from NVMe."
echo "4. SSH in with: ssh $USERNAME@$HOSTNAME_SET.local"
echo
echo "After first boot, recommended hardening:"
echo "  sudo sed -i 's/^#\?PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config"
echo "  Set: PasswordAuthentication no"
echo "  Then: sudo systemctl restart ssh"
echo
echo "If SSH asks for a password:"
echo "- verify /home/$USERNAME/.ssh/authorized_keys exists"
echo "- verify permissions are 700 on .ssh and 600 on authorized_keys"
echo "- verify ownership matches $USERNAME"
