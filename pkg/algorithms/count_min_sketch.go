package algorithms

import (
	"hash/fnv"
	"math"
)

type CountMinSketch struct {
	Height uint64
	Width  uint64
	Table  [][]uint64
}

func NewCMS(height, width uint64) *CountMinSketch {
	table := make([][]uint64, height)
	for i := range table {
		table[i] = make([]uint64, width)
	}

	return &CountMinSketch{
		Height: height,
		Width:  width,
		Table:  table,
	}
}

func (s *CountMinSketch) hash(item string, seed uint64) uint64 {
	h := fnv.New64a()
	h.Write([]byte(item))
	return h.Sum64() ^ seed
}

func (s *CountMinSketch) Insert(data string) {
	var row uint64
	for row = 0; row < s.Height; row++ {
		col := s.hash(data, row) % s.Width
		s.Table[row][col]++
	}
}

func (s *CountMinSketch) Estimate(data string) uint64 {
	estimation := uint64(math.MaxUint64)

	var row uint64
	for row = 0; row < s.Height; row++ {
		col := s.hash(data, row) % s.Width
		freq := s.Table[row][col]

		estimation = min(estimation, freq)
	}

	return estimation
}
