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

func GetLocations(k0 uint64, k1 uint64, numHash uint64, data []byte) []uint64 {
	// @todo
	return []uint64{}
}

// TestLocations returns true if all locations are set in the BitSet,
// false otherwise.
func TestLocations(b *BitSet, locations []uint64) bool {
	for i := 0; i < len(locations); i++ {
		if !b.Test(locations[i]) {
			return false
		}
	}
	return true
}

func SetLocations(b *BitSet, locations []uint64) *BitSet {
	for _, loc := range locations {
		b.Set(loc)
	}
	return b
}
