package algorithms

import "hash/fnv"

// hashPair returns two independent-ish 32-bit hashes of item for
// Kirsch–Mitzenmacher double hashing: index_i = h1 + i*h2.
// h2 is forced odd so rows never collapse onto the same column.
func hashPair(item string) (uint64, uint64) {
	h := fnv.New64a()
	h.Write([]byte(item))
	sum := h.Sum64()
	return sum & 0xffffffff, (sum >> 32) | 1
}
