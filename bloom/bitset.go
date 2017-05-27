package bloom

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
	return &BitSet{
		numBits: max(1, numBits),
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

// Length returns the number of bits of the BitSet
func (b *BitSet) Length() uint64 {
	return b.numBits
}

// Bytes returns the raw bitset data
// - can be passed into From to recreate the BitSet
func (b *BitSet) Bytes() []uint64 {
	return b.data
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

func max(x uint64, y uint64) uint64 {
	if x > y {
		return x
	}
	return y
}
