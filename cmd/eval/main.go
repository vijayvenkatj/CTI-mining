// Command eval replays a frozen pulse stream (JSONL) through the real edge
// generation code and TRIEST, and writes the evaluation CSVs + summary.md.
// Everything except perf.csv is deterministic for a given dataset and -seed.
package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/vijayvenkatj/cti-miner/pkg/algorithms"
	edgegen "github.com/vijayvenkatj/cti-miner/pkg/edge-gen"
	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

const inf = uint64(math.MaxUint64)

var (
	dataPath   = flag.String("data", "eval/data/pulses.jsonl", "frozen pulse stream (JSONL)")
	outDir     = flag.String("out", "eval/results", "output directory")
	thresholds = flag.String("thresholds", "2,5,10,25,50,100,250,1000,inf", "CMS thresholds to sweep (inf = baseline)")
	reservoirs = flag.String("reservoirs", "1000,5000,10000,25000,50000", "TRIEST reservoir sizes M")
	runs       = flag.Int("runs", 10, "TRIEST runs per M")
	seed       = flag.Uint64("seed", 42, "base seed for TRIEST")
	prefixes   = flag.String("prefixes", "0.1,0.25,0.5,0.75,1.0", "stream prefixes for scaling")
	retain     = flag.Float64("retain", 0.9, "min triangle retention for the chosen threshold")
	cmsDepth   = flag.Uint64("cms-depth", 4, "pipeline CMS depth (production: 4)")
	cmsWidth   = flag.Uint64("cms-width", 1<<10, "pipeline CMS width (production: 1024)")
	bloomBits  = flag.Uint64("bloom-bits", 1<<24, "pipeline Bloom bits (production: 2^24)")
	bloomK     = flag.Uint64("bloom-k", 4, "pipeline Bloom hashes (production: 4)")
	triestM    = flag.Int("m", 5000, "TRIEST M used in the perf runs (production: 5000)")
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	pulses := load(*dataPath)
	must(os.MkdirAll(*outDir, 0o755))
	nInd := 0
	for _, p := range pulses {
		nInd += len(p.Indicators)
	}
	log.Printf("loaded %d pulses, %d indicator occurrences", len(pulses), nInd)

	perf := newCSV("perf.csv", "experiment", "prefix", "threshold", "pulses", "edges", "peak_live_heap_mb", "seconds", "pulses_per_s", "edges_per_s")
	defer perf.Flush()

	// Exp 1 + 2: threshold sweep (graph reduction + correlation preservation).
	ths := parseThresholds(*thresholds)
	if ths[len(ths)-1] != inf {
		ths = append(ths, inf) // baseline is always needed
	}
	type row struct {
		th    uint64
		build build
		stats graphStats
	}
	rows := make([]row, len(ths))
	for i, th := range ths {
		b := buildGraph(pulses, th)
		rows[i] = row{th, b, analyse(b.edges)}
		log.Printf("threshold %s: %d edges, %d triangles", fmtTh(th), len(b.edges), rows[i].stats.Triangles)
	}
	base := rows[len(rows)-1]
	tc := newCSV("threshold.csv", "threshold", "indicators_filtered", "candidate_edges", "unique_edges", "edge_reduction_pct",
		"nodes", "index_entries", "exact_triangles", "triangle_retention_pct", "components", "groups_ge3", "largest_component")
	chosen := base
	perfBy := map[uint64]perfResult{}
	for _, r := range rows {
		ret := pct(r.stats.Triangles, base.stats.Triangles)
		if chosen.th == inf && r.th != inf && ret >= *retain*100 {
			chosen = r // smallest threshold that keeps enough structure
		}
		tc.row(fmtTh(r.th), r.build.filtered, r.build.candidates, len(r.build.edges), reduction(len(r.build.edges), len(base.build.edges)),
			r.stats.Nodes, r.build.indexEntries, r.stats.Triangles, ret, r.stats.Components, r.stats.Groups, r.stats.Largest)
		p := measure(pulses, r.th)
		perfBy[r.th] = p
		perf.row("threshold", 1.0, fmtTh(r.th), len(pulses), p.edges, p.heapMB, p.seconds, rate(len(pulses), p.seconds), rate(p.edges, p.seconds))
	}
	tc.Flush()
	chosenNote := fmt.Sprintf("smallest with >= %.0f%% triangle retention", *retain*100)
	if chosen.th == inf && len(rows) > 1 {
		// Nothing met -retain: fall back to the least aggressive finite threshold.
		chosen = rows[len(rows)-2]
		chosenNote = fmt.Sprintf("NO threshold reached %.0f%% retention; fell back to the largest finite threshold", *retain*100)
	}
	log.Printf("chosen threshold: %s (%s)", fmtTh(chosen.th), chosenNote)

	// Exp 3: TRIEST accuracy vs reservoir size, on baseline and chosen graphs.
	trc := newCSV("triest.csv", "graph", "M", "run", "seed", "estimate", "exact", "abs_err", "rel_err_pct")
	summ := map[string]map[int]acc{}
	for _, g := range []row{base, chosen} {
		name := "baseline"
		if g.th != inf {
			name = "threshold_" + fmtTh(g.th)
		}
		summ[name] = map[int]acc{}
		exact := g.stats.Triangles
		for _, m := range parseInts(*reservoirs) {
			if m >= len(g.build.edges) {
				log.Printf("skip M=%d on %s: M >= |E|=%d (TRIEST is exact there)", m, name, len(g.build.edges))
				continue
			}
			var sum, sq, ape float64
			for r := range *runs {
				s := *seed + uint64(r)
				t := algorithms.NewTriestWithSeed(m, s)
				for _, e := range g.build.edges {
					t.Insert(e)
				}
				est := t.Estimate()
				abs := math.Abs(float64(est - exact))
				rel := pctF(abs, float64(exact))
				trc.row(name, m, r, s, est, exact, abs, rel)
				sum += float64(est)
				sq += float64(est) * float64(est)
				ape += rel
			}
			n := float64(*runs)
			mean := sum / n
			summ[name][m] = acc{mean, math.Sqrt(max(0, sq/n-mean*mean)), pctF(math.Abs(mean-float64(exact)), float64(exact)), ape / n}
		}
	}
	trc.Flush()
	tsc := newCSV("triest_summary.csv", "graph", "M", "exact", "mean", "std", "rel_err_of_mean_pct", "mape_pct")
	for _, g := range []row{base, chosen} {
		name := "baseline"
		if g.th != inf {
			name = "threshold_" + fmtTh(g.th)
		}
		for _, m := range parseInts(*reservoirs) {
			if a, ok := summ[name][m]; ok {
				tsc.row(name, m, g.stats.Triangles, a.mean, a.std, a.relErr, a.mape)
			}
		}
	}
	tsc.Flush()

	// Exp 4: scaling over real stream prefixes (no synthetic duplication).
	sc := newCSV("scaling.csv", "prefix", "pulses", "config", "threshold", "unique_edges", "index_entries", "exact_triangles", "cms_bytes", "bloom_bytes", "reservoir_edges")
	for _, f := range parseFloats(*prefixes) {
		sub := pulses[:int(f*float64(len(pulses)))]
		for _, cfg := range []struct {
			name string
			th   uint64
		}{{"baseline", inf}, {"proposed", chosen.th}} {
			b := buildGraph(sub, cfg.th)
			sc.row(f, len(sub), cfg.name, fmtTh(cfg.th), len(b.edges), b.indexEntries, analyse(b.edges).Triangles,
				*cmsDepth**cmsWidth*8, *bloomBits/8, min(*triestM, len(b.edges)))
			p := measure(sub, cfg.th)
			perf.row("scaling_"+cfg.name, f, fmtTh(cfg.th), len(sub), p.edges, p.heapMB, p.seconds, rate(len(sub), p.seconds), rate(p.edges, p.seconds))
		}
	}
	sc.Flush()

	// Exp 5: CMS correctness and the effect of width/depth (legacy vs double hashing).
	exactCnt := map[string]uint64{}
	var stream []string
	for _, p := range pulses {
		for _, ind := range p.Indicators {
			exactCnt[ind.Indicator]++
			stream = append(stream, ind.Indicator)
		}
	}
	cc := newCSV("cms.csv", "hash", "width", "depth", "distinct_keys", "mean_abs_err", "max_err", "overestimate_rate_pct", "violations", "false_filter_rate_pct")
	for _, hash := range []string{"legacy", "double"} {
		for _, w := range []uint64{256, 1024, 4096, 16384} {
			for _, d := range []uint64{1, 2, 4, 8} {
				est := cmsEstimator(hash, d, w, stream)
				var sumErr, maxErr float64
				var over, viol, below, falseFilt int
				for k, c := range exactCnt {
					e := est(k)
					if e < c {
						viol++
					}
					if e > c {
						over++
					}
					err := float64(e) - float64(c)
					sumErr += err
					maxErr = max(maxErr, err)
					if chosen.th != inf && c < chosen.th {
						below++
						if e >= chosen.th {
							falseFilt++
						}
					}
				}
				cc.row(hash, w, d, len(exactCnt), sumErr/float64(len(exactCnt)), maxErr, pct(over, len(exactCnt)), viol, ratio(falseFilt, below))
			}
		}
	}
	cc.Flush()

	// Exp 6: edges silently dropped by Bloom dedup false positives.
	bc := newCSV("bloom.csv", "threshold", "bloom_bits", "k", "exact_unique_edges", "bloom_emitted_edges", "dropped_pct")
	for _, r := range []row{base, chosen} {
		for _, bits := range []uint64{1 << 16, 1 << 20, *bloomBits} { // 2^16 = old production size
			eg := newGen(r.th, algorithms.NewBloomFilter(bits, *bloomK))
			n := 0
			for _, p := range pulses {
				must(eg.ProcessPulse(p, func(resources.Edge) error { n++; return nil }))
			}
			bc.row(fmtTh(r.th), bits, *bloomK, len(r.build.edges), n, reduction(n, len(r.build.edges)))
		}
	}
	bc.Flush()

	writeSummary(len(pulses), base.build, chosen.build, base.stats, chosen.stats, chosen.th, perfBy[inf], perfBy[chosen.th], summ, chosenNote)
	log.Printf("results written to %s", *outDir)
}

type build struct {
	edges                              []resources.Edge // unique, in stream order (what TRIEST sees)
	candidates, filtered, indexEntries int
}

func newGen(th uint64, bloom *algorithms.BloomFilter) *edgegen.EdgeGenerator {
	return edgegen.NewEdgeGenerator(bloom, algorithms.NewCMS(*cmsDepth, *cmsWidth), &resources.IndicatorIndex{}, th, nil, nil)
}

// buildGraph runs the production ProcessPulse with exact dedup instead of
// Bloom, so graph metrics are not polluted by Bloom false positives (Exp 6).
func buildGraph(pulses []resources.Pulse, th uint64) build {
	eg := newGen(th, nil)
	seen := map[resources.Edge]struct{}{}
	var b build
	for _, p := range pulses {
		must(eg.ProcessPulse(p, func(e resources.Edge) error {
			b.candidates++
			if _, ok := seen[e]; !ok {
				seen[e] = struct{}{}
				b.edges = append(b.edges, e)
			}
			return nil
		}))
	}
	distinct := map[string]struct{}{}
	for _, p := range pulses {
		for _, ind := range p.Indicators {
			distinct[ind.Indicator] = struct{}{}
		}
	}
	for k := range distinct {
		if eg.CMS.Estimate(k) >= th {
			b.filtered++
		}
	}
	b.indexEntries = eg.IndicatorIndex.Entries()
	return b
}

type perfResult struct {
	edges           int
	heapMB, seconds float64
}

// measure runs the production config (Bloom dedup + TRIEST(M)) twice: once
// untouched for timing, once with periodic GC to sample peak live heap.
// ponytail: in-process with baseline-heap subtraction; run configs as separate
// processes if the numbers look contaminated.
func measure(pulses []resources.Pulse, th uint64) perfResult {
	run := func(sample bool) (int, float64, time.Duration) {
		runtime.GC()
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		baseHeap := ms.HeapAlloc
		var peak uint64
		eg := newGen(th, algorithms.NewBloomFilter(*bloomBits, *bloomK))
		t := algorithms.NewTriestWithSeed(*triestM, *seed)
		edges := 0
		start := time.Now()
		for i, p := range pulses {
			must(eg.ProcessPulse(p, func(e resources.Edge) error { edges++; t.Insert(e); return nil }))
			if sample && (i%1000 == 999 || i == len(pulses)-1) {
				runtime.GC()
				runtime.ReadMemStats(&ms)
				peak = max(peak, ms.HeapAlloc)
			}
		}
		el := time.Since(start)
		runtime.KeepAlive(eg)
		runtime.KeepAlive(t)
		return edges, float64(peak-min(peak, baseHeap)) / (1 << 20), el
	}
	edges, _, el := run(false)
	_, heap, _ := run(true)
	return perfResult{edges, heap, el.Seconds()}
}

// cmsEstimator builds a sketch over stream; "legacy" reproduces the old
// fnv^row hashing so the write-up can show why it was replaced.
func cmsEstimator(hash string, d, w uint64, stream []string) func(string) uint64 {
	if hash == "double" {
		s := algorithms.NewCMS(d, w)
		for _, k := range stream {
			s.Insert(k)
		}
		return s.Estimate
	}
	h := func(k string, row uint64) uint64 {
		f := fnv.New64a()
		f.Write([]byte(k))
		return (f.Sum64() ^ row) % w
	}
	table := make([][]uint64, d)
	for i := range table {
		table[i] = make([]uint64, w)
	}
	for _, k := range stream {
		for r := range d {
			table[r][h(k, r)]++
		}
	}
	return func(k string) uint64 {
		e := inf
		for r := range d {
			e = min(e, table[r][h(k, r)])
		}
		return e
	}
}

type acc struct{ mean, std, relErr, mape float64 }

// writeSummary emits the final Baseline vs CMS+TRIEST table and the
// data-driven conclusion sentence. TRIEST error is taken at -m.
func writeSummary(n int, bb, cb build, bs, cs graphStats, th uint64, bp, cp perfResult, summ map[string]map[int]acc, chosenNote string) {
	f, err := os.Create(filepath.Join(*outDir, "summary.md"))
	must(err)
	defer f.Close()
	errAt := func(name string) string {
		if a, ok := summ[name][*triestM]; ok {
			return fmt.Sprintf("%.0f ± %.0f (%.2f%%)", a.mean, a.std, a.mape)
		}
		return "exact (M >= |E|)"
	}
	cname := "threshold_" + fmtTh(th)
	edgeRed := reduction(len(cb.edges), len(bb.edges))
	memRed := 0.0
	if bp.heapMB > 0 {
		memRed = (bp.heapMB - cp.heapMB) / bp.heapMB * 100
	}
	ret := pct(cs.Triangles, bs.Triangles)
	w := func(format string, a ...any) { fmt.Fprintf(f, format, a...) }
	w("# Evaluation summary\n\n")
	w("Dataset: %d pulses. CMS %dx%d, Bloom %d bits/k=%d, TRIEST M=%d, %d runs, seed %d. Chosen threshold: %s (%s).\n\n",
		n, *cmsDepth, *cmsWidth, *bloomBits, *bloomK, *triestM, *runs, *seed, fmtTh(th), chosenNote)
	w("| Metric | Baseline | CMS + TRIEST |\n|---|---:|---:|\n")
	w("| Pulses processed | %d | %d |\n", n, n)
	w("| Candidate edges | %d | %d |\n", bb.candidates, cb.candidates)
	w("| Unique edges | %d | %d |\n", len(bb.edges), len(cb.edges))
	w("| Edge reduction | - | %.2f%% |\n", edgeRed)
	w("| Indicators filtered | %d | %d |\n", bb.filtered, cb.filtered)
	w("| Index entries | %d | %d |\n", bb.indexEntries, cb.indexEntries)
	w("| Peak live heap | %.2f MB | %.2f MB |\n", bp.heapMB, cp.heapMB)
	w("| Throughput | %.0f pulses/s | %.0f pulses/s |\n", rate(n, bp.seconds), rate(n, cp.seconds))
	w("| Exact triangles | %d | %d |\n", bs.Triangles, cs.Triangles)
	w("| TRIEST estimate (M=%d, mean ± std, MAPE) | %s | %s |\n", *triestM, errAt("baseline"), errAt(cname))
	w("| Triangle retention | - | %.2f%% |\n", ret)
	w("| Correlated groups (>=3 pulses) | %d | %d |\n\n", bs.Groups, cs.Groups)
	mape := "n/a (exact regime)"
	if a, ok := summ[cname][*triestM]; ok {
		mape = fmt.Sprintf("%.2f%%", a.mape)
	}
	w("The proposed architecture reduced graph construction by %.2f%% and peak memory by %.2f%%, while retaining %.2f%% of the baseline triangle structure, with a TRIEST estimation error (MAPE at M=%d) of %s.\n\n",
		edgeRed, memRed, ret, *triestM, mape)
	w("Throughput and memory are in-process (Kafka excluded) and vary between runs; all other numbers are deterministic.\n")
}

func load(path string) []resources.Pulse {
	f, err := os.Open(path)
	must(err)
	defer f.Close()
	var out []resources.Pulse
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<30)
	for sc.Scan() {
		var p resources.Pulse
		must(json.Unmarshal(sc.Bytes(), &p))
		out = append(out, p)
	}
	must(sc.Err())
	return out
}

type csvFile struct {
	*csv.Writer
	f *os.File
}

func newCSV(name string, header ...string) *csvFile {
	f, err := os.Create(filepath.Join(*outDir, name))
	must(err)
	c := &csvFile{csv.NewWriter(f), f}
	must(c.Write(header))
	return c
}

func (c *csvFile) row(vals ...any) {
	rec := make([]string, len(vals))
	for i, v := range vals {
		switch v := v.(type) {
		case float64:
			rec[i] = strconv.FormatFloat(v, 'f', 4, 64)
		default:
			rec[i] = fmt.Sprint(v)
		}
	}
	must(c.Write(rec))
}

func parseThresholds(s string) []uint64 {
	var out []uint64
	for _, p := range strings.Split(s, ",") {
		if p == "inf" {
			out = append(out, inf)
			continue
		}
		v, err := strconv.ParseUint(p, 10, 64)
		must(err)
		out = append(out, v)
	}
	return out
}

func parseInts(s string) []int {
	var out []int
	for _, p := range strings.Split(s, ",") {
		v, err := strconv.Atoi(p)
		must(err)
		out = append(out, v)
	}
	return out
}

func parseFloats(s string) []float64 {
	var out []float64
	for _, p := range strings.Split(s, ",") {
		v, err := strconv.ParseFloat(p, 64)
		must(err)
		out = append(out, v)
	}
	return out
}

func fmtTh(t uint64) string {
	if t == inf {
		return "inf"
	}
	return strconv.FormatUint(t, 10)
}

func pct(a, b int) float64 { return pctF(float64(a), float64(b)) }

func pctF(a, b float64) float64 {
	if b == 0 {
		return 100 // nothing to lose: e.g. 0 baseline triangles
	}
	return a / b * 100
}

func ratio(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b) * 100
}

func reduction(got, base int) float64 {
	if base == 0 {
		return 0
	}
	return float64(base-got) / float64(base) * 100
}

func rate(n int, s float64) float64 {
	if s == 0 {
		return 0
	}
	return float64(n) / s
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
