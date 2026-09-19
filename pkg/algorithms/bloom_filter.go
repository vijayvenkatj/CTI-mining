package algorithms

import "hash/fnv"

type BloomFilter struct {
	N    uint64
	K    uint64
	Bits []bool
}

func NewBloomFilter(nBits, nHashes uint64) *BloomFilter {
	return &BloomFilter{
		N:    nBits,
		K:    nHashes,
		Bits: make([]bool, nBits),
	}
}

func (b *BloomFilter) hash(item string, seed uint64) uint64 {
	h := fnv.New64a()
	h.Write([]byte(item))
	return h.Sum64() ^ seed
}

func (b *BloomFilter) Insert(data string) {
	var i uint64
	for i = 0; i < b.K; i++ {
		hash := b.hash(data, i) % b.N
		b.Bits[hash] = true
	}
}

func (b *BloomFilter) Contains(data string) bool {
	var i uint64
	for i = 0; i < b.K; i++ {
		hash := b.hash(data, i) % b.N
		if !b.Bits[hash] {
			return false
		}
	}
	return true
}
