package drbg

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"

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

	seed.value = make([]byte, SeedLength)
	_, err := rand.Read(seed.value)
	if err != nil {
		return nil, err
	}

	return seed, nil
}

func (s *Seed) UnmarshalBinary(data []byte) error {
	if len(data) < SeedLength {
		return errors.New("Invalid DRBG seed. Too few bytes.")
	}
	s.value = data
	return nil
}

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
		return errors.New("invlaid drbg seed length")
	}
	return nil
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
