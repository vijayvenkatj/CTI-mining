package algorithms

type BloomFilter struct {
	N    uint64
	K    uint64
	Bits []uint64 // bitset: N bits packed into N/64 words
}

func NewBloomFilter(nBits, nHashes uint64) *BloomFilter {
	return &BloomFilter{
		N:    nBits,
		K:    nHashes,
		Bits: make([]uint64, (nBits+63)/64),
	}
}

func (b *BloomFilter) Insert(data string) {
	h1, h2 := hashPair(data)
	var i uint64
	for i = 0; i < b.K; i++ {
		hash := (h1 + i*h2) % b.N
		b.Bits[hash/64] |= 1 << (hash % 64)
	}
}

func (b *BloomFilter) Contains(data string) bool {
	h1, h2 := hashPair(data)
	var i uint64
	for i = 0; i < b.K; i++ {
		hash := (h1 + i*h2) % b.N
		if b.Bits[hash/64]&(1<<(hash%64)) == 0 {
			return false
		}
	}
	return true
}
