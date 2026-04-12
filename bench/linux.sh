#!/usr/bin/env bash
# Run the harness inside OrbStack cray-vm. Host paths auto-mount.
#
# Usage: bench/linux.sh <baseline-binary> <candidate-binary> <label> [iterations]
set -euo pipefail

if [[ $# -lt 3 ]]; then
  echo "usage: $0 <baseline> <candidate> <label> [iterations]" >&2
  exit 2
fi

BASELINE=$(cd "$(dirname "$1")" && pwd)/$(basename "$1")
CANDIDATE=$(cd "$(dirname "$2")" && pwd)/$(basename "$2")
LABEL="$3"
ITERATIONS="${4:-2000}"

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
VM="${CMDPERF_VM:-cray-vm}"

if ! command -v orb >/dev/null 2>&1; then
  echo "orb CLI not found in PATH" >&2
  exit 1
fi

echo "=> Running Linux harness inside OrbStack VM: $VM"
orb -m "$VM" bash "$SCRIPT_DIR/run.sh" "$BASELINE" "$CANDIDATE" "$LABEL" "$ITERATIONS"
