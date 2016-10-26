package drbg

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/dchest/siphash"
)

const SeedLength = 16 + siphash.Size

// Initial state for a HashDrbg.
// - SipHash-2-4 keys: key0 and key1
// - 8 byte nonce (initialization vector)
type Seed struct {
	value []byte // Key (first 16 bytes) + InitVec (8 bytes)
}

func NewSeed() (*Seed, error) {
	seed := &Seed{}

	seed.value = make([]byte, 16+siphash.Size)
	_, err := rand.Read(seed.value)
	if err != nil {
		return nil, err
	}

	return seed, nil
}

func ImportSeed(b []byte) (*Seed, error) {
	seed := &Seed{}
	if len(b) < (16 + siphash.Size) {
		return nil, fmt.Errorf("Invalid DRBG seed. Too few bytes %v", len(b))
	}
	seed.value = b
	return seed, nil
}

func (s *Seed) Export() []byte {
	return s.value[:]
}

func (s *Seed) Key() []byte {
	return s.value[:16]
}

func (s *Seed) KeyUint128() (uint64, uint64) {
	s1, _ := binary.Uvarint(s.value[0:8])
	s2, _ := binary.Uvarint(s.value[8:16])
	return s1, s2
}

func (s *Seed) InitVec() []byte {
	return s.value[16:]
}
