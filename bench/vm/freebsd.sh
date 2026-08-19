#!/usr/bin/env bash
# Run the cmdperf harness inside a local FreeBSD arm64 VM (QEMU + HVF).
#
# More reliable than GitHub-hosted runners for absolute numbers: dedicated
# cores, no noisy neighbors, hypervisor-accelerated (not emulated).
#
# First invocation builds a golden image (base cloud image + go/bash/git,
# ~10 min). Every benchmark run after that boots a fresh copy-on-write
# overlay on the golden image in ~30s.
#
# Usage: bench/vm/freebsd.sh <baseline-ref> <candidate-ref> <label> [iterations]
#   refs are git refs (e.g. main, HEAD, a branch); binaries are built inside
#   the VM so they are native FreeBSD arm64.
#
# Requirements: qemu (brew install qemu), an ssh key in ~/.ssh, ~3GB disk.
set -euo pipefail

if [[ $# -lt 3 ]]; then
  echo "usage: $0 <baseline-ref> <candidate-ref> <label> [iterations]" >&2
  exit 2
fi

BASELINE_REF="$1"
CANDIDATE_REF="$2"
LABEL="$3"
ITERATIONS="${4:-2000}"

RELEASE="14.3-RELEASE"
IMAGE_NAME="FreeBSD-${RELEASE}-arm64-aarch64-BASIC-CLOUDINIT-ufs"
IMAGE_URL="https://download.freebsd.org/releases/VM-IMAGES/${RELEASE}/aarch64/Latest/${IMAGE_NAME}.qcow2.xz"

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/../.." && pwd)
CACHE_DIR="$SCRIPT_DIR/cache"
GOLDEN="$CACHE_DIR/golden-freebsd-${RELEASE}.qcow2"
OUT_ROOT="$REPO_ROOT/bench/results/$LABEL"
SSH_PORT="${CMDPERF_VM_SSH_PORT:-10422}"
SSH_OPTS=(-p "$SSH_PORT" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR root@localhost)
FIRMWARE="$(dirname "$(command -v qemu-system-aarch64)")/../share/qemu/edk2-aarch64-code.fd"
QEMU_PID="$CACHE_DIR/qemu-$LABEL.pid"

mkdir -p "$CACHE_DIR"

boot_vm() { # $1: disk image
  qemu-system-aarch64 \
    -machine virt -accel hvf -cpu host -smp 4 -m 2048 \
    -bios "$FIRMWARE" \
    -drive if=virtio,file="$1",format=qcow2 \
    -drive if=virtio,file="$CACHE_DIR/seed.iso",format=raw,readonly=on \
    -netdev user,id=n0,hostfwd=tcp:127.0.0.1:$SSH_PORT-:22 -device virtio-net-pci,netdev=n0 \
    -display none -daemonize -pidfile "$QEMU_PID"
}

stop_vm() { [[ -f "$QEMU_PID" ]] && kill "$(cat "$QEMU_PID")" 2>/dev/null; rm -f "$QEMU_PID"; }

wait_ssh() {
  echo -n "=> Waiting for ssh"
  for _ in $(seq 1 60); do
    if ssh "${SSH_OPTS[@]}" true 2>/dev/null; then echo " up"; return 0; fi
    echo -n "."; sleep 5
  done
  echo " TIMEOUT" >&2; return 1
}

# --- NoCloud seed for FreeBSD's nuageinit(8). It ignores cloud-config
# modules like `packages`, but executes shell-script user-data as root,
# so provisioning is a script: create the user, install toolchain. ---
make_seed() {
  local pubkey seed_dir
  pubkey=$(cat "$(ls "$HOME"/.ssh/id_*.pub | head -1)")
  seed_dir=$(mktemp -d)
  printf 'instance-id: cmdperf\nlocal-hostname: cmdperf-bsd\n' > "$seed_dir/meta-data"
  cat > "$seed_dir/user-data" <<EOF
#!/bin/sh
# Minimal boot-time setup: root ssh only. Package installation happens
# over ssh from the host (visible, retryable) rather than blind at boot.
mkdir -p /root/.ssh
echo '$pubkey' > /root/.ssh/authorized_keys
chmod 700 /root/.ssh
chmod 600 /root/.ssh/authorized_keys
sysrc sshd_enable=YES
echo 'PermitRootLogin prohibit-password' >> /etc/ssh/sshd_config
service sshd restart || service sshd start
# The image ships a firstboot freebsd-update that can reboot mid-provision;
# disable it so golden builds are deterministic.
pkg delete -fy firstboot-freebsd-update 2>/dev/null || true
sysrc -x firstboot_freebsd_update_enable 2>/dev/null || true
EOF
  rm -f "$CACHE_DIR/seed.iso"
  hdiutil makehybrid -quiet -iso -joliet -default-volume-name cidata \
    -o "$CACHE_DIR/seed.iso" "$seed_dir"
  rm -rf "$seed_dir"
}

# --- Golden image: base + packages, built once ---
build_golden() {
  if [[ ! -f "$CACHE_DIR/$IMAGE_NAME.qcow2" ]]; then
    echo "=> Downloading $IMAGE_NAME"
    curl -L --fail -o "$CACHE_DIR/$IMAGE_NAME.qcow2.xz" "$IMAGE_URL"
    xz -d "$CACHE_DIR/$IMAGE_NAME.qcow2.xz"
  fi

  echo "=> Building golden image (one-time, ~10 min)"
  qemu-img create -q -f qcow2 -F qcow2 -b "$CACHE_DIR/$IMAGE_NAME.qcow2" "$GOLDEN.tmp" 20G
  boot_vm "$GOLDEN.tmp"
  wait_ssh

  echo "=> Installing go, bash, git"
  for attempt in 1 2 3; do
    if ssh "${SSH_OPTS[@]}" 'env ASSUME_ALWAYS_YES=yes pkg install -y go bash git'; then break; fi
    [[ "$attempt" == 3 ]] && { echo "pkg install failed" >&2; exit 1; }
    sleep 15
  done
  ssh "${SSH_OPTS[@]}" 'command -v go >/dev/null && command -v bash >/dev/null && command -v git >/dev/null'

  echo "=> Shutting down golden VM"
  ssh "${SSH_OPTS[@]}" 'shutdown -p now' || true
  for _ in $(seq 1 24); do
    kill -0 "$(cat "$QEMU_PID" 2>/dev/null)" 2>/dev/null || break
    sleep 5
  done
  stop_vm
  mv "$GOLDEN.tmp" "$GOLDEN"
}

trap 'stop_vm' EXIT

make_seed
[[ -f "$GOLDEN" ]] || build_golden

# --- Benchmark run: fresh overlay on golden, boots in ~30s ---
OVERLAY="$CACHE_DIR/run-$LABEL.qcow2"
rm -f "$OVERLAY"
qemu-img create -q -f qcow2 -F qcow2 -b "$GOLDEN" "$OVERLAY" 20G

echo "=> Booting benchmark VM (ssh port $SSH_PORT)"
boot_vm "$OVERLAY"
wait_ssh

echo "=> Syncing repository"
git -C "$REPO_ROOT" bundle create -q "$CACHE_DIR/repo.bundle" "$BASELINE_REF" "$CANDIDATE_REF"
scp -q -P "$SSH_PORT" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
  "$CACHE_DIR/repo.bundle" root@localhost:/tmp/repo.bundle

BASELINE_SHA=$(git -C "$REPO_ROOT" rev-parse "$BASELINE_REF")
CANDIDATE_SHA=$(git -C "$REPO_ROOT" rev-parse "$CANDIDATE_REF")

echo "=> Building and benchmarking (iterations=$ITERATIONS)"
ssh "${SSH_OPTS[@]}" BASELINE_SHA="$BASELINE_SHA" CANDIDATE_SHA="$CANDIDATE_SHA" \
  LABEL="$LABEL" ITERATIONS="$ITERATIONS" 'sh -s' <<'REMOTE'
set -e
rm -rf /tmp/src && mkdir /tmp/src && cd /tmp/src
git init -q
git fetch -q /tmp/repo.bundle '+refs/*:refs/bundle/*'
git checkout -q "$BASELINE_SHA"  && go build -o /tmp/baseline  ./cmd/cmdperf
git checkout -q "$CANDIDATE_SHA" && go build -o /tmp/candidate ./cmd/cmdperf
bash bench/run.sh /tmp/baseline /tmp/candidate "$LABEL" "$ITERATIONS"
REMOTE

mkdir -p "$OUT_ROOT"
scp -q -r -P "$SSH_PORT" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
  "root@localhost:/tmp/src/bench/results/$LABEL/*" "$OUT_ROOT/"

echo
go run "$REPO_ROOT/bench/compare" "$OUT_ROOT/baseline" "$OUT_ROOT/candidate"
