package drbg

import (
	"encoding/binary"
	"errors"
	"hash"
	"sync"

	"github.com/dchest/siphash"
)

// HashDrbg is a CSDRBG based off of SipHash-2-4 in OFB mode.
type HashDrbg struct {
	mu   sync.Mutex
	seed *Seed
	sip  hash.Hash64
	ofb  [siphash.Size]byte
	//ctr uint64
}

// NewHashDrbg creates a deterministic random number generator from a provided Seed.
func NewHashDrbg(seed *Seed) (*HashDrbg, error) {
	d := &HashDrbg{}
	if seed == nil {
		newSeed, seedErr := NewSeed()
		if seedErr != nil {
			return nil, seedErr
		}
		d.seed = newSeed
	} else {
		d.seed = seed
	}
	//d.ctr = 0
	d.sip = siphash.New(d.seed.Key())
	copy(d.ofb[:], d.seed.InitVec())

	return d, nil
}

/********************
 * PUBLIC METHODS
 ********************/

// Next returns the next 8 byte DRBG block.
func (d *HashDrbg) Next() []byte {
	d.mu.Lock()
	d.sip.Write(d.ofb[:])
	copy(d.ofb[:], d.sip.Sum(nil))
	ret := make([]byte, siphash.Size)
	copy(ret, d.ofb[:])
	d.mu.Unlock()
	return ret
}

// RandomUint32 provides the next block of the random number generator as an integer.
func (d *HashDrbg) RandomUint32() uint32 {
	block := d.Next()
	ret := binary.LittleEndian.Uint32(block)
	return ret
}

// RandomUint64 provides the next block of the random number generator as a long integer.
func (d *HashDrbg) RandomUint64() uint64 {
	block := d.Next()
	ret := binary.LittleEndian.Uint64(block)
	return ret
}

// FillBytes fills a byte slice from the random number sequence.
func (d *HashDrbg) FillBytes(b []byte) {
	randBytes := d.Next()

	for i := 0; i < len(b); i++ {
		b[i] = randBytes[0]
		if len(randBytes) < 2 {
			randBytes = d.Next()
		} else {
			randBytes = randBytes[1:]
		}
	}
}

// Overlay is a static method for xoring a byte array with the deterministic
// random sequence generated from a provided seed.
func Overlay(seed, data []byte) error {
	if len(seed) < SeedLength {
		return errors.New("invalid seed provided")
	}
	s := Seed{}
	s.UnmarshalBinary(seed)
	d, err := NewHashDrbg(&s)
	if err != nil {
		return err
	}

	dbytes := d.Next()
	for i := 0; i < len(data); i++ {
		data[i] ^= dbytes[0]
		if len(dbytes) < 2 {
			dbytes = d.Next()
		} else {
			dbytes = dbytes[1:]
		}
	}
	return nil
}
