# Evaluation

Reproduces every number and plot for the CMS + TRIEST evaluation from a frozen
pulse stream. The harness (`cmd/eval`) calls the production
`EdgeGenerator.ProcessPulse` and `Triest`, with Kafka replaced by a file replay.

## 1. Export the dataset (once)

With the compose Postgres running (`docker compose up -d postgres`) and populated by ingestion:

```bash
docker exec -i postgres psql -U cti -d ctiminer -At < eval/export.sql > eval/data/pulses.jsonl
shasum -a 256 eval/data/pulses.jsonl > eval/data/pulses.sha256
wc -l eval/data/pulses.jsonl   # record the pulse count in the report
```

Commit `pulses.sha256` (the `.jsonl` is gitignored), and keep a copy of the
`.jsonl` so others can rerun the evaluation on the same data.

## 2. Run

```bash
pip install matplotlib
eval/run.sh                    # defaults: see `go run ./cmd/eval -h`
eval/run.sh -runs 20 -retain 0.8
```

Outputs go to `eval/results/`:

| File | Experiment |
|---|---|
| `threshold.csv` | 1 + 2: edges, edge reduction, index size, exact triangles, retention, components per CMS threshold (`inf` = baseline) |
| `triest.csv`, `triest_summary.csv` | 3: TRIEST estimate vs exact for each reservoir M x `-runs` seeds (mean, std, MAPE) |
| `scaling.csv` | 4: baseline vs proposed on real stream prefixes (no synthetic duplication) |
| `perf.csv` | 4: peak live heap, seconds, pulses/s, edges/s (the only non-deterministic file) |
| `cms.csv` | 5: CMS error grid, width x depth x {legacy fnv^row, double hashing}; `violations` must be 0 |
| `bloom.csv` | 6: unique edges silently dropped by Bloom dedup false positives |
| `summary.md` | Final baseline vs CMS+TRIEST table and conclusion sentence |
| `*.png` | Plots 1-6 |

The baseline is the same pipeline with threshold = infinity, so a filtered graph
is always a subgraph of the baseline. The "chosen" threshold is the smallest one
with triangle retention >= `-retain`. If none reaches it, the summary says so.

## Notes for the write-up

- Graph metrics use exact edge dedup. Bloom loss is reported separately in `bloom.csv`.
- Threshold `t` means an indicator links at most `t-1` pulses before it is
  dropped. At `t=2` no edges can form.
- Memory is the peak live heap of edge-gen + TRIEST(M=`-m`) in-process, and
  throughput excludes Kafka.
- Code fixes made for the evaluation: Bloom resized 2^16 -> 2^24 bits and
  packed into a bitset (2 MB; the old size dropped >90% of edges, see
  `bloom.csv` 65536 rows), CMS/Bloom double hashing (previously
  `fnv ^ row`, so depth > 1 had no effect, see `cms.csv` legacy rows), and
  filtered indicators are now removed from the indicator index.
