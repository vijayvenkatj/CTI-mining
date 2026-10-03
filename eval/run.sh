#!/usr/bin/env bash
# Reproduce every evaluation result from the frozen dataset.
# Usage: eval/run.sh [extra cmd/eval flags...]
set -euo pipefail
cd "$(dirname "$0")/.."

DATA=eval/data/pulses.jsonl
[ -f "$DATA" ] || { echo "missing $DATA - see eval/README.md (Export)"; exit 1; }
if [ -f eval/data/pulses.sha256 ]; then
  shasum -a 256 -c eval/data/pulses.sha256
else
  echo "warning: no eval/data/pulses.sha256, dataset is not pinned"
fi

go test ./...
go run ./cmd/eval -data "$DATA" -out eval/results "$@"
python3 eval/plot.py eval/results
