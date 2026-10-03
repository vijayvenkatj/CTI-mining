package algorithms

import "math"

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

func (s *CountMinSketch) Insert(data string) {
	h1, h2 := hashPair(data)
	var row uint64
	for row = 0; row < s.Height; row++ {
		col := (h1 + row*h2) % s.Width
		s.Table[row][col]++
	}
}

func (s *CountMinSketch) Estimate(data string) uint64 {
	estimation := uint64(math.MaxUint64)

	h1, h2 := hashPair(data)
	var row uint64
	for row = 0; row < s.Height; row++ {
		col := (h1 + row*h2) % s.Width
		freq := s.Table[row][col]

		estimation = min(estimation, freq)
	}

	return estimation
}
