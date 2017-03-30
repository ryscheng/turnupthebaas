package drbg

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/dchest/siphash"
)

// SeedLength is the number of bytes a drbg seed takes in memory.
const SeedLength = 16 + siphash.Size

// Seed holds the state of a deterministic HashDrbg.
// - SipHash-2-4 keys: key0 and key1
// - 8 byte nonce (initialization vector)
type Seed struct {
	value []byte // Key (first 16 bytes) + InitVec (8 bytes)
}

// NewSeed creates a new Seed
func NewSeed() (*Seed, error) {
	seed := &Seed{}

	seed.value = make([]byte, SeedLength)
	_, err := rand.Read(seed.value)
	if err != nil {
		return nil, err
	}

	return seed, nil
}

// UnmarshalBinary reconstructs a Seed from a binary implementation, implementing the interface.
func (s *Seed) UnmarshalBinary(data []byte) error {
	if len(data) < SeedLength {
		return errors.New("invalid DRBG seed. Too few bytes")
	}
	s.value = data
	return nil
}

// MarshalBinary creates a byte array representation of a Seed.
func (s *Seed) MarshalBinary() ([]byte, error) {
	return s.value[:], nil
}

// MarshalJSON serializes the seed to JSON
// Implements the json.Marshaller interface
func (s *Seed) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.value)
}

// UnmarshalJSON restores the seed from a JSON representation.
func (s *Seed) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &s.value); err != nil {
		return err
	}
	if len(s.value) != SeedLength {
		return errors.New("invalid drbg seed length")
	}
	return nil
}

// Key provides the byte representation of the underlying key for the Seed
func (s *Seed) Key() []byte {
	return s.value[:16]
}

// KeyUint128 provides a representation of the underlying key as two 64bit ints.
func (s *Seed) KeyUint128() (uint64, uint64) {
	s1, _ := binary.Uvarint(s.value[0:8])
	s2, _ := binary.Uvarint(s.value[8:16])
	return s1, s2
}

// InitVec provides the initialization vector of the seed.
func (s *Seed) InitVec() []byte {
	return s.value[16:]
}
