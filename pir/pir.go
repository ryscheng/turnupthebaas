package pir

import (
	"fmt"
	"github.com/ilyak/bitvec"
	"unsafe"
)

type Database interface {
	Read(uint32) []byte
	Length() uint32
}

type PIRServer struct {
	Database
}

func (p PIRServer) String() string {
	return fmt.Sprintf("PIRServer with %d keys", p.Database.Length())
}

const wordSize = int(unsafe.Sizeof(uintptr(0)))

// Assumes that cells are word aligned & multiples of 8 bytes.
// Extracted from https://golang.org/src/crypto/cipher/xor.go
func (p *PIRServer) Read(requests []bitvec.BitVec) [][]byte {
	n := len(requests)
	out := make([][]byte, n)
	cell := p.Database.Read(0)
	cellLength := len(cell)

	for j := 0; j < n; j++ {
		out[j] = make([]byte, cellLength)
	}

	words := cellLength / wordSize

	for i := uint32(0); i < p.Database.Length(); i++ {
		cellPtr := p.Database.Read(i)
		cell := *(*[]uintptr)(unsafe.Pointer(&cellPtr))
		for j := 0; j < n; j++ {
			if requests[j].IsSet(int(i)) {
				dest := *(*[]uintptr)(unsafe.Pointer(&out[j]))
				for k := 0; k < words; k++ {
					dest[i] ^= cell[i]
				}
			}
		}
	}
	return out
}
