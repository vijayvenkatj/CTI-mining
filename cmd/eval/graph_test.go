package main

import (
	"fmt"
	"testing"

	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

func clique(prefix string, n int) []resources.Edge {
	var es []resources.Edge
	for i := range n {
		for j := i + 1; j < n; j++ {
			es = append(es, resources.MakeEdge(fmt.Sprint(prefix, i), fmt.Sprint(prefix, j)))
		}
	}
	return es
}

func TestAnalyse(t *testing.T) {
	if got := analyse(clique("a", 4)).Triangles; got != 4 {
		t.Fatalf("K4 = %d, want 4", got)
	}
	if got := analyse(clique("a", 5)).Triangles; got != 10 {
		t.Fatalf("K5 = %d, want 10", got)
	}
	s := analyse(append(append(clique("a", 3), clique("b", 4)...), resources.MakeEdge("x", "y")))
	if s.Components != 3 || s.Groups != 2 || s.Largest != 4 || s.Nodes != 9 {
		t.Fatalf("got %+v", s)
	}
}

func TestFilteredGraphIsSubgraph(t *testing.T) {
	ind := func(vals ...string) []resources.Indicator {
		var out []resources.Indicator
		for _, v := range vals {
			out = append(out, resources.Indicator{Indicator: v})
		}
		return out
	}
	// "common" appears in every pulse; "x"/"y" link a few pulses.
	var pulses []resources.Pulse
	for i := range 8 {
		vals := []string{"common"}
		if i%2 == 0 {
			vals = append(vals, "x")
		}
		if i < 3 {
			vals = append(vals, "y")
		}
		pulses = append(pulses, resources.Pulse{ID: fmt.Sprint("p", i), Indicators: ind(vals...)})
	}
	base := buildGraph(pulses, inf)
	filt := buildGraph(pulses, 3)
	in := map[resources.Edge]bool{}
	for _, e := range base.edges {
		in[e] = true
	}
	for _, e := range filt.edges {
		if !in[e] {
			t.Fatalf("filtered edge %v not in baseline", e)
		}
	}
	if len(filt.edges) >= len(base.edges) || filt.filtered == 0 || filt.indexEntries >= base.indexEntries {
		t.Fatalf("filtering had no effect: base %d edges/%d entries, filtered %d edges/%d entries",
			len(base.edges), base.indexEntries, len(filt.edges), filt.indexEntries)
	}
}
