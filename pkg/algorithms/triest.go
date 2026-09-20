package algorithms

import (
	"math/rand/v2"

	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

type Triest struct {
	T   int
	M   int
	tau int

	Edges  []resources.Edge
	AdjMap map[string]map[string]struct{}
}

func NewTreist(m int) *Triest {
	return &Triest{
		T:   0,
		M:   m,
		tau: 0,

		Edges:  []resources.Edge{},
		AdjMap: make(map[string]map[string]struct{}),
	}
}

func (t *Triest) Insert(e resources.Edge) {
	t.T++

	if len(t.Edges) < t.M {
		t.tau += t.triangles(e)
		t.addEdge(e)
		return
	}

	// There is M/T chance of an edge being selected
	if rand.Float64() >= float64(t.M)/float64(t.T) {
		return
	}

	idx := rand.IntN(len(t.Edges))
	removed := t.Edges[idx]

	t.tau -= t.triangles(removed)
	t.removeEdge(removed)

	t.tau += t.triangles(e)
	t.addEdge(e, idx)
}

func (t *Triest) Estimate() int {
	if t.T <= t.M {
		return t.tau
	}

	T := float64(t.T)
	M := float64(t.M)
	factor := (T * (T - 1) * (T - 2)) / (M * (M - 1) * (M - 2))

	return int(factor * float64(t.tau))
}

func (t *Triest) triangles(e resources.Edge) int {
	result := 0
	for neighbour := range t.AdjMap[e.Source] {
		if _, exists := t.AdjMap[e.Target][neighbour]; exists {
			result++
		}
	}
	return result
}

func (t *Triest) addEdge(e resources.Edge, i ...int) {
	if t.AdjMap[e.Source] == nil {
		t.AdjMap[e.Source] = make(map[string]struct{})
	}
	if t.AdjMap[e.Target] == nil {
		t.AdjMap[e.Target] = make(map[string]struct{})
	}

	t.AdjMap[e.Source][e.Target] = struct{}{}
	t.AdjMap[e.Target][e.Source] = struct{}{}

	if len(i) > 0 {
		t.Edges[i[0]] = e
		return
	}
	t.Edges = append(t.Edges, e)
}

func (t *Triest) removeEdge(e resources.Edge) {
	delete(t.AdjMap[e.Source], e.Target)
	delete(t.AdjMap[e.Target], e.Source)
}
