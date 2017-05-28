package bloom

import (
	"math/rand"
	"testing"
)

func TestEstimateParameters(t *testing.T) {
	l, h := EstimateParameters(1000000, 0.01)
	if l <= 1000000 {
		t.Errorf("Expected numBits=%d to be greater than 1000000", l)
	}
	if h <= 4 {
		t.Errorf("Expected numHash=%d to be greater than 4", h)
	}
}

func TestGetLocations(t *testing.T) {
	numHash := uint64(7)
	data := []byte("test")
	key := make([]byte, 16)
	n, err := rand.Read(key)
	if n < 16 || err != nil {
		t.Errorf("Error generating random key")
	}
	locs := GetLocations(key, numHash, data)
	if uint64(len(locs)) != numHash {
		t.Errorf("Mismatched locations: len=%v, expected=%v", len(locs), numHash)
	}
	for _, v := range locs {
		if v == 0 {
			t.Errorf("Location is 0")
		}
	}
}

func TestSetCheckLocations(t *testing.T) {
	b := NewBitSet(32)
	locs := []uint64{2, 4, 6, 8}
	offLocs := []uint64{2, 1}
	if CheckLocations(b, offLocs) {
		t.Errorf("Fresh BitSet has bits set")
	}
	b = SetLocations(b, locs)
	if !CheckLocations(b, locs) {
		t.Errorf("BitSet doesn't see what was previously set")
	}
	if !CheckLocations(b, locs[:2]) {
		t.Errorf("BitSet doesn't see what was previously set")
	}
	if CheckLocations(b, offLocs) {
		t.Errorf("BitSet improperly set")
	}

}
