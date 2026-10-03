"""Plot the eval CSVs. Usage: python3 eval/plot.py [results_dir]"""
import csv, sys
from collections import defaultdict
from pathlib import Path

import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt

d = Path(sys.argv[1] if len(sys.argv) > 1 else "eval/results")
rows = lambda name: list(csv.DictReader(open(d / name)))
f = lambda r, k: float(r[k])

def save(fig, name):
    fig.tight_layout()
    fig.savefig(d / name, dpi=150)
    plt.close(fig)

# Thresholds on a categorical axis so "inf" (baseline) fits.
th = rows("threshold.csv")
labels = [r["threshold"] for r in th]
x = range(len(labels))

fig, ax = plt.subplots(figsize=(7, 4))
ax.plot(x, [f(r, "edge_reduction_pct") for r in th], "o-", label="Edge reduction %")
ax.plot(x, [f(r, "triangle_retention_pct") for r in th], "s-", label="Triangle retention %")
ax.set(xticks=list(x), xticklabels=labels, xlabel="CMS threshold", ylabel="%", ylim=(-5, 105),
       title="Graph reduction vs correlation preservation")
ax.legend(); ax.grid(alpha=.3)
save(fig, "1_threshold_tradeoff.png")

perf = [r for r in rows("perf.csv") if r["experiment"] == "threshold"]
for col, ylabel, name in [("peak_live_heap_mb", "Peak live heap (MB)", "2_threshold_memory.png"),
                          ("pulses_per_s", "Pulses / s", "3_threshold_throughput.png")]:
    fig, ax = plt.subplots(figsize=(7, 4))
    ax.bar(x, [f(r, col) for r in perf])
    ax.set(xticks=list(x), xticklabels=labels, xlabel="CMS threshold", ylabel=ylabel)
    ax.grid(alpha=.3, axis="y")
    save(fig, name)

fig, ax = plt.subplots(figsize=(7, 4))
by = defaultdict(list)
for r in rows("triest_summary.csv"):
    by[r["graph"]].append(r)
for g, rs in by.items():
    ax.errorbar([f(r, "M") for r in rs], [f(r, "mape_pct") for r in rs], fmt="o-", capsize=3, label=g)
ax.set(xscale="log", xlabel="Reservoir size M (edges)", ylabel="Mean abs. % error",
       title="TRIEST accuracy vs memory budget")
ax.legend(); ax.grid(alpha=.3)
save(fig, "4_triest_error.png")

sc = rows("scaling.csv")
sp = {(r["experiment"], r["prefix"]): r for r in rows("perf.csv") if r["experiment"].startswith("scaling_")}
fig, (a1, a2) = plt.subplots(1, 2, figsize=(11, 4))
for cfg in ("baseline", "proposed"):
    rs = [r for r in sc if r["config"] == cfg]
    px = [int(r["pulses"]) for r in rs]
    a1.plot(px, [int(r["unique_edges"]) for r in rs], "o-", label=cfg)
    a2.plot(px, [f(sp[("scaling_" + cfg, r["prefix"])], "peak_live_heap_mb") for r in rs], "o-", label=cfg)
a1.set(xlabel="Pulses (real stream prefix)", ylabel="Unique edges", title="Graph growth")
a2.set(xlabel="Pulses (real stream prefix)", ylabel="Peak live heap (MB)", title="Memory growth")
for a in (a1, a2):
    a.legend(); a.grid(alpha=.3)
save(fig, "5_scaling.png")

cms = rows("cms.csv")
widths = sorted({int(r["width"]) for r in cms}); depths = sorted({int(r["depth"]) for r in cms})
fig, axes = plt.subplots(1, 2, figsize=(11, 4), sharey=True)
for ax, h in zip(axes, ("legacy", "double")):
    grid = [[next(f(r, "mean_abs_err") for r in cms if r["hash"] == h and int(r["width"]) == w and int(r["depth"]) == dp)
             for w in widths] for dp in depths]
    im = ax.imshow(grid, cmap="viridis_r", aspect="auto")
    top = max(map(max, grid))
    for i, rw in enumerate(grid):
        for j, v in enumerate(rw):
            ax.text(j, i, f"{v:.2f}", ha="center", va="center", color="w" if v > top / 2 else "k", fontsize=8)
    ax.set(xticks=range(len(widths)), xticklabels=widths, yticks=range(len(depths)), yticklabels=depths,
           xlabel="width", title=f"CMS mean abs error ({h} hash)")
    fig.colorbar(im, ax=ax)
axes[0].set_ylabel("depth")
save(fig, "6_cms_error.png")
print("plots written to", d)
