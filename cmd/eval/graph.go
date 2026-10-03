package main

import (
	"slices"

	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

type graphStats struct {
	Nodes, Triangles, Components, Groups, Largest int
}

// analyse computes exact graph statistics. Edges must be unique.
// Groups = connected components with >= 3 pulses (correlated pulse groups).
func analyse(edges []resources.Edge) graphStats {
	ids := map[string]int{}
	id := func(s string) int {
		if v, ok := ids[s]; ok {
			return v
		}
		ids[s] = len(ids)
		return len(ids) - 1
	}
	type pair struct{ a, b int }
	pairs := make([]pair, len(edges))
	for i, e := range edges {
		pairs[i] = pair{id(e.Source), id(e.Target)}
	}
	n := len(ids)

	deg := make([]int, n)
	for _, p := range pairs {
		deg[p.a]++
		deg[p.b]++
	}
	// Forward algorithm: orient each edge from lower to higher (degree, id) rank,
	// then count common out-neighbours. O(m^1.5).
	less := func(u, v int) bool { return deg[u] < deg[v] || (deg[u] == deg[v] && u < v) }
	out := make([][]int, n)
	for _, p := range pairs {
		if less(p.a, p.b) {
			out[p.a] = append(out[p.a], p.b)
		} else {
			out[p.b] = append(out[p.b], p.a)
		}
	}
	for _, o := range out {
		slices.Sort(o)
	}
	triangles := 0
	for u := range n {
		for _, v := range out[u] {
			a, b := out[u], out[v]
			for i, j := 0, 0; i < len(a) && j < len(b); {
				switch {
				case a[i] < b[j]:
					i++
				case a[i] > b[j]:
					j++
				default:
					triangles++
					i++
					j++
				}
			}
		}
	}

	parent := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	find := func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	for _, p := range pairs {
		parent[find(p.a)] = find(p.b)
	}
	size := map[int]int{}
	for i := range n {
		size[find(i)]++
	}
	s := graphStats{Nodes: n, Triangles: triangles, Components: len(size)}
	for _, c := range size {
		if c >= 3 {
			s.Groups++
		}
		s.Largest = max(s.Largest, c)
	}
	return s
}
