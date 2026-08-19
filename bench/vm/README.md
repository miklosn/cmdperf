# Local VM benchmarking

Authoritative-numbers path: local VMs on Apple Silicon beat GitHub-hosted
runners for absolute measurements — dedicated cores, no noisy neighbors,
hypervisor-accelerated (HVF), never instruction-emulated.

GitHub workflows (`benchmark.yaml`) remain the cross-platform regression
gate on PRs; use these scripts when the actual numbers matter.

| Target | Script | Backend |
|---|---|---|
| Linux (arm64) | `../linux.sh` | OrbStack (`CMDPERF_VM` selects the machine) |
| FreeBSD (arm64) | `freebsd.sh` | QEMU + HVF, official cloud image |

`freebsd.sh` takes git refs, builds both inside the VM, runs the ABAB
harness, and prints the comparison:

```bash
bench/vm/freebsd.sh main my-branch fbsd-local 2000
```

First run downloads the FreeBSD cloud image into `cache/` (~500MB,
gitignored). Each run boots a fresh copy-on-write overlay, so results
always start from a clean VM.

Caveat: arm64 guests measure the BSD kernel on arm under a hypervisor;
absolute floors differ from bare-metal x86, but A/B comparisons and jitter
characteristics transfer well.

Requirements: `brew install qemu`, an ssh public key in `~/.ssh`.
