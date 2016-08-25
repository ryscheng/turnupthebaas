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

func NewHashDrbg(seed *Seed) *HashDrbg {
	d := &HashDrbg{}
	if seed == nil {
		d.seed, _ = NewSeed()
	} else {
		d.seed = seed
	}
	//d.ctr = 0
	d.sip = siphash.New(seed.Key)
	copy(d.ofb[:], seed.InitVec)

	return d
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

func (d *HashDrbg) FillBytes(b *[]byte) {
	randBytes := d.Next()

	for i := 0; i < len(*b); i++ {
		(*b)[i] = randBytes[0]
		if len(randBytes) < 2 {
			randBytes = d.Next()
		} else {
			randBytes = randBytes[1:]
		}
	}
}
