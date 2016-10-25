package drbg

import (
	"encoding/binary"
	"github.com/dchest/siphash"
	"hash"
)

// HashDrbg is a CSDRBG based off of SipHash-2-4 in OFB mode.
type HashDrbg struct {
	seed *Seed
	sip  hash.Hash64
	ofb  [siphash.Size]byte
	//ctr uint64
}

func NewHashDrbg(seed *Seed) (*HashDrbg, error) {
	d := &HashDrbg{}
	if seed == nil {
		newSeed, seedErr := NewSeed()
		if seedErr != nil {
			return nil, seedErr
		} else {
			d.seed = newSeed
		}
	} else {
		d.seed = seed
	}
	//d.ctr = 0
	d.sip = siphash.New(d.seed.Key())
	copy(d.ofb[:], d.seed.InitVec())

	return d, nil
}

// NextBlock returns the next 8 byte DRBG block.
func (d *HashDrbg) Next() []byte {
	d.sip.Write(d.ofb[:])
	copy(d.ofb[:], d.sip.Sum(nil))
	ret := make([]byte, siphash.Size)
	copy(ret, d.ofb[:])
	return ret
}

func (d *HashDrbg) RandomUint32() uint32 {
	block := d.Next()
	ret := binary.LittleEndian.Uint32(block)
	return ret
}

func (d *HashDrbg) RandomUint64() uint64 {
	block := d.Next()
	ret := binary.LittleEndian.Uint64(block)
	return ret
}

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
