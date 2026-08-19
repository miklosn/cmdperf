#!/usr/bin/env bash
# Run the cmdperf harness on a dedicated GCP spot VM.
#
# Quieter than shared CI runners: you own the whole instance for the run.
# Spot pricing keeps a ~10-minute benchmark at a few cents; the instance is
# always deleted afterwards (trap), and spot termination auto-deletes.
#
# Usage: bench/gcp/run.sh <baseline-ref> <candidate-ref> <label> [machine-type] [iterations]
#   machine-type: c4d-standard-4 (x86, default) or c4a-standard-4 (arm)
set -euo pipefail

if [[ $# -lt 3 ]]; then
  echo "usage: $0 <baseline-ref> <candidate-ref> <label> [machine-type] [iterations]" >&2
  exit 2
fi

BASELINE_REF="$1"
CANDIDATE_REF="$2"
LABEL="$3"
MACHINE_TYPE="${4:-c4d-standard-4}"
ITERATIONS="${5:-2000}"

PROJECT="${CMDPERF_GCP_PROJECT:-clamav-eval-cray}"
ZONE="${CMDPERF_GCP_ZONE:-europe-west1-b}"
INSTANCE="cmdperf-bench-$LABEL"

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/../.." && pwd)
OUT_ROOT="$REPO_ROOT/bench/results/$LABEL"
GC=(gcloud --project "$PROJECT")

cleanup() {
  echo "=> Deleting instance $INSTANCE"
  "${GC[@]}" compute instances delete "$INSTANCE" --zone "$ZONE" --quiet 2>/dev/null || true
}
trap cleanup EXIT

echo "=> Creating spot instance $INSTANCE ($MACHINE_TYPE, $ZONE)"
"${GC[@]}" compute instances create "$INSTANCE" \
  --zone "$ZONE" \
  --machine-type "$MACHINE_TYPE" \
  --provisioning-model=SPOT \
  --instance-termination-action=DELETE \
  --network="${CMDPERF_GCP_NETWORK:-vm-creation-network}" \
  --subnet="${CMDPERF_GCP_SUBNET:-vm-creation-network}" \
  --image-family="$([[ "$MACHINE_TYPE" == c4a-* || "$MACHINE_TYPE" == t2a-* ]] && echo debian-12-arm64 || echo debian-12)" \
  --image-project=debian-cloud \
  --boot-disk-type=hyperdisk-balanced --boot-disk-size=20GB \
  --tags=cmdperf-bench \
  --labels=purpose=cmdperf-bench

echo -n "=> Waiting for ssh"
for _ in $(seq 1 30); do
  if "${GC[@]}" compute ssh "$INSTANCE" --zone "$ZONE" --command true 2>/dev/null; then
    echo " up"; break
  fi
  echo -n "."; sleep 10
done
"${GC[@]}" compute ssh "$INSTANCE" --zone "$ZONE" --command true

GO_VERSION="1.24.13"
echo "=> Installing go $GO_VERSION, git"
"${GC[@]}" compute ssh "$INSTANCE" --zone "$ZONE" --command \
  "sudo DEBIAN_FRONTEND=noninteractive apt-get -qq update >/dev/null && sudo DEBIAN_FRONTEND=noninteractive apt-get -qq install -y git >/dev/null && curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-\$(dpkg --print-architecture).tar.gz | sudo tar -C /usr/local -xz"

echo "=> Syncing repository"
BASELINE_SHA=$(git -C "$REPO_ROOT" rev-parse "$BASELINE_REF")
CANDIDATE_SHA=$(git -C "$REPO_ROOT" rev-parse "$CANDIDATE_REF")
BUNDLE=$(mktemp -t cmdperf-bundle)
git -C "$REPO_ROOT" bundle create -q "$BUNDLE" "$BASELINE_REF" "$CANDIDATE_REF"
"${GC[@]}" compute scp "$BUNDLE" "$INSTANCE:/tmp/repo.bundle" --zone "$ZONE" --quiet
rm -f "$BUNDLE"

echo "=> Building and benchmarking (iterations=$ITERATIONS)"
"${GC[@]}" compute ssh "$INSTANCE" --zone "$ZONE" --command "
set -e
export PATH=\$PATH:/usr/local/go/bin
rm -rf /tmp/src && mkdir /tmp/src && cd /tmp/src
git init -q
git fetch -q /tmp/repo.bundle '+refs/*:refs/bundle/*'
git checkout -q $BASELINE_SHA  && go build -o /tmp/baseline  ./cmd/cmdperf
git checkout -q $CANDIDATE_SHA && go build -o /tmp/candidate ./cmd/cmdperf
bash bench/run.sh /tmp/baseline /tmp/candidate '$LABEL' '$ITERATIONS'
"

mkdir -p "$OUT_ROOT"
"${GC[@]}" compute scp --recurse \
  "$INSTANCE:/tmp/src/bench/results/$LABEL/*" "$OUT_ROOT/" --zone "$ZONE" --quiet

echo
go run "$REPO_ROOT/bench/compare" "$OUT_ROOT/baseline" "$OUT_ROOT/candidate"
