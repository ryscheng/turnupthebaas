package bloom

import (
	"encoding/binary"
	"math"

	"github.com/dchest/siphash"
)

// EstimateParameters estimates requirements for the number of bits and hash functions,
// to use in the Bloom Filter.
// Computed from the number of items to add and the false positive rate
func EstimateParameters(numItems uint64, fpRate float64) (numBits uint64, numHash uint64) {
	l := uint64(math.Ceil(-1 * float64(numItems) * math.Log(fpRate) / math.Pow(math.Log(2), 2)))
	h := uint64(math.Ceil(math.Log(2) * float64(l) / float64(numItems)))
	return l, h
}

// GetLocations will generate a list of numHash bit locations
// using a cryptographic hash function keyed on a 16-byte key
// and the specified data
func GetLocations(key []byte, numHash uint64, data []byte) []uint64 {
	var loc uint64
	buf := make([]byte, 8)
	result := make([]uint64, 0)

	h := siphash.New(key)
	h.Write(data)
	for i := uint64(0); i < numHash; i++ {
		binary.PutUvarint(buf, i)
		h.Write(buf)
		loc = h.Sum64()
		result = append(result, loc)
	}
	return result
}

// CheckLocations returns true if all locations are set in the BitSet,
// false otherwise.
func CheckLocations(b *BitSet, locations []uint64) bool {
	for i := 0; i < len(locations); i++ {
		if !b.Test(locations[i] % b.Length()) {
			return false
		}
	}
	return true
}

// SetLocations will set all specified bits to 1 in the bitset
func SetLocations(b *BitSet, locations []uint64) *BitSet {
	for _, loc := range locations {
		b.Set(loc % b.Length())
	}
	return b
}
