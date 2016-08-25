package drbg

import (
	"crypto/rand"
	"github.com/dchest/siphash"
)

const SeedLength = 16 + siphash.Size

// Initial state for a HashDrbg.
// - SipHash-2-4 keys: key0 and key1
// - 8 byte nonce (initialization vector)
type Seed struct {
	Key     []byte //16 bytes
	InitVec []byte //8 bytes
}

func NewSeed() (*Seed, error) {
	seed := &Seed{}

	seed.Key = make([]byte, 16)
	_, err := rand.Read(seed.Key)
	if err != nil {
		return nil, err
	}
	seed.InitVec = make([]byte, siphash.Size)
	_, err = rand.Read(seed.InitVec)
	if err != nil {
		return nil, err
	}

	return seed, nil
}
