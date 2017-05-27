package bloom

import (
	"math"
)

// EstimateParameters estimates requirements for the number of bits and hash functions,
// to use in the Bloom Filter.
// Computed from the number of items to add and the false positive rate
func EstimateParameters(numItems uint64, fpRate float64) (numBits uint64, numHash uint64) {
	l := uint64(math.Ceil(-1 * float64(numItems) * math.Log(fpRate) / math.Pow(math.Log(2), 2)))
	h := uint64(math.Ceil(math.Log(2) * float64(l) / float64(numItems)))
	return l, h
}

// Filter is a Bloom filter of n items
type Filter struct {
	numBits uint64
	numHash uint64
	data    []uint64
}

// NewFilter creates a new Bloom filter with numBits bits and numHash hashing functions
// Parameters are forced to be at least 1
func NewFilter(numBits uint64, numHash uint64) *Filter {
	return &Filter{
		numBits: max(1, numBits), 
		numHash: max(1, numHash), 
		data: bitset.New(m)
	}
}

func max(x, y uint) uint {
	if x > y {
		return x
	}
	return y
}
