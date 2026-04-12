#!/usr/bin/env bash
# Drive a comparison benchmark of two cmdperf binaries.
#
# Usage: bench/run.sh <baseline-binary> <candidate-binary> <label> [iterations]
#
# Writes CSVs to bench/results/<label>/{baseline,candidate}/<workload-slug>.csv
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
WORKLOADS="$SCRIPT_DIR/workloads.txt"
OUT_ROOT="$SCRIPT_DIR/results/$LABEL"
mkdir -p "$OUT_ROOT/baseline" "$OUT_ROOT/candidate"

if [[ ! -x "$BASELINE" ]]; then echo "baseline binary not executable: $BASELINE" >&2; exit 1; fi
if [[ ! -x "$CANDIDATE" ]]; then echo "candidate binary not executable: $CANDIDATE" >&2; exit 1; fi

# Optional CPU pinning on Linux
PIN=()
if command -v taskset >/dev/null 2>&1; then
  PIN=(taskset -c 0)
fi

slug() {
  echo "$1" | tr -c '[:alnum:]' '_' | sed 's/__*/_/g' | sed 's/^_//;s/_$//'
}

run_one() {
  local bin="$1" outdir="$2" workload="$3"
  local s; s=$(slug "$workload")
  "${PIN[@]}" "$bin" -n "$ITERATIONS" --csv "$outdir/$s.csv" "$workload" >/dev/null 2>&1 || {
    echo "  ! $(basename "$bin") failed on: $workload" >&2
  }
}

echo "=> Running harness: label=$LABEL iterations=$ITERATIONS"
echo "   baseline:  $BASELINE"
echo "   candidate: $CANDIDATE"
echo "   workloads: $WORKLOADS"
echo

# ABAB interleaving: run baseline, candidate, baseline, candidate per workload
while IFS= read -r workload; do
  [[ -z "$workload" || "$workload" == \#* ]] && continue
  echo "-- workload: $workload"
  for rep in 1 2; do
    run_one "$BASELINE" "$OUT_ROOT/baseline" "$workload"
    run_one "$CANDIDATE" "$OUT_ROOT/candidate" "$workload"
  done
done < "$WORKLOADS"

echo
echo "Done. Results in: $OUT_ROOT"
