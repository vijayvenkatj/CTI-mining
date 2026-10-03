package algorithms

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

func TestCMSNeverUnderestimates(t *testing.T) {
	s := NewCMS(4, 64) // tiny width forces collisions
	exact := map[string]uint64{}
	r := rand.New(rand.NewPCG(1, 1))
	for range 10000 {
		k := fmt.Sprint(r.IntN(500))
		s.Insert(k)
		exact[k]++
	}
	for k, c := range exact {
		if e := s.Estimate(k); e < c {
			t.Fatalf("%s: estimate %d < exact %d", k, e, c)
		}
	}
}

func TestCMSDepthUsesDistinctColumns(t *testing.T) {
	// Regression for fnv^row hashing: keys colliding in row 0 must not collide in every row.
	h1a, h2a := hashPair("a")
	for i := range 1000 {
		k := fmt.Sprint(i)
		h1, h2 := hashPair(k)
		if h1%1024 == h1a%1024 && (h1+h2)%1024 == (h1a+h2a)%1024 && (h1+2*h2)%1024 == (h1a+2*h2a)%1024 {
			t.Fatalf("%q collides with \"a\" in rows 0-2", k)
		}
	}
}

func k4() []resources.Edge {
	var es []resources.Edge
	n := []string{"a", "b", "c", "d"}
	for i := range n {
		for j := i + 1; j < len(n); j++ {
			es = append(es, resources.MakeEdge(n[i], n[j]))
		}
	}
	return es
}

func TestTriestExactWhenReservoirFits(t *testing.T) {
	tr := NewTriestWithSeed(100, 7)
	for _, e := range k4() {
		tr.Insert(e)
	}
	if got := tr.Estimate(); got != 4 {
		t.Fatalf("K4 triangles = %d, want 4", got)
	}
}

func TestTriestSeedIsReproducible(t *testing.T) {
	var es []resources.Edge
	for i := range 60 {
		for j := i + 1; j < 60; j += 3 {
			es = append(es, resources.MakeEdge(fmt.Sprint(i), fmt.Sprint(j)))
		}
	}
	run := func() int {
		tr := NewTriestWithSeed(50, 9)
		for _, e := range es {
			tr.Insert(e)
		}
		return tr.Estimate()
	}
	if a, b := run(), run(); a != b {
		t.Fatalf("same seed gave %d and %d", a, b)
	}
}

func TestBloomNoFalseNegatives(t *testing.T) {
	b := NewBloomFilter(1000, 4) // not a multiple of 64 on purpose
	for i := range 500 {
		b.Insert(fmt.Sprint(i))
	}
	for i := range 500 {
		if !b.Contains(fmt.Sprint(i)) {
			t.Fatalf("%d inserted but not found", i)
		}
	}
	if NewBloomFilter(1000, 4).Contains("x") {
		t.Fatal("empty filter claims membership")
	}
}
