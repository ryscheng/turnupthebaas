package bloom

import (
	"log"
	"math"
)

// log2WordSize is lg(wordSize)
const log2WordSize = uint64(6)

// the wordSize of a bit set
const wordSize = uint64(64)

// BitSet is a set of bits
type BitSet struct {
	numBits uint64
	data    []uint64
}

// NewBitSet creates a new BitSet with numBits bits
// Parameters are forced to be at least 1
func NewBitSet(numBits uint64) *BitSet {
	// Golang cannot support slices where len(s) is greater than the default integer size
	if numBits > math.MaxUint32 {
		log.Printf("NewBitSet cannot handle numBits=%d\n", numBits)
		return nil
	}
	return &BitSet{
		numBits: numBits,
		data:    make([]uint64, wordsNeeded(numBits)),
	}
}

// From is a constructor used to create a BitSet from an array of integers
func From(numBits uint64, buf []uint64) *BitSet {
	return &BitSet{
		numBits: numBits,
		data:    buf,
	}
}

// Equal tests the equvalence of two BitSets.
// False if they are of different sizes, otherwise true
// only if all the same bits are set
func Equal(a *BitSet, b *BitSet) bool {
	if a == nil || b == nil {
		return false
	}
	if a.Length() != b.Length() {
		return false
	}
	if a.Length() == 0 { // if they have both length == 0, then could have nil set
		return true
	}
	// testing for equality shoud not transform the bitset (no call to safeSet)

	aBytes := a.Bytes()
	bBytes := b.Bytes()
	for i, v := range aBytes {
		if bBytes[i] != v {
			return false
		}
	}
	return true

}

// Length returns the number of bits of the BitSet
func (b *BitSet) Length() uint64 {
	return b.numBits
}

// Bytes returns the raw bitset data
// - can be passed into From to recreate the BitSet
func (b *BitSet) Bytes() []uint64 {
	return b.data[:]
}

// Test whether bit i is set.
func (b *BitSet) Test(i uint64) bool {
	if i >= b.numBits {
		return false
	}
	return b.data[i>>log2WordSize]&(1<<(i&(wordSize-1))) != 0
}

// Set bit i to 1
func (b *BitSet) Set(i uint64) bool {
	if i >= b.numBits {
		return false
	}
	b.data[i>>log2WordSize] |= 1 << (i & (wordSize - 1))
	return true
}

// Clear bit i to 0
func (b *BitSet) Clear(i uint64) bool {
	if i >= b.numBits {
		return false
	}
	b.data[i>>log2WordSize] &^= 1 << (i & (wordSize - 1))
	return true
}

// SetTo sets bit i to value
func (b *BitSet) SetTo(i uint64, value bool) bool {
	if value {
		return b.Set(i)
	}
	return b.Clear(i)
}

// wordsNeeded calculates the number of words needed for i bits
func wordsNeeded(i uint64) uint64 {
	if i > (Cap() - wordSize + 1) {
		return (Cap() >> log2WordSize)
	}
	return (i + (wordSize - 1)) >> log2WordSize
}

// Cap returns the total possible capacity, or number of bits
func Cap() uint64 {
	return ^uint64(0)
}
