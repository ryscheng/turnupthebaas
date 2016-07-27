package pir

import (
	"github.com/ilyak/bitvec"
	"math/rand"
	"testing"
)

const cellSize = 1024
const cellCount = 1024
const batchSize = 8

type boringDB struct {
	Data [][]byte
}

func (b boringDB) Read(n uint32) []byte {
	return b.Data[n]
}

func (b boringDB) Length() uint32 {
	return uint32(cellCount)
}

func BenchmarkRead(b *testing.B) {
	//Make Database
	var db boringDB
	db.Data = make([][]byte, cellCount)
	for i := 0; i < cellCount; i++ {
		db.Data[i] = make([]byte, cellSize)
	}
	server := PIRServer{db}

	//Make testVector
	testVector := make([]bitvec.BitVec, batchSize)
	for i := 0; i < batchSize; i++ {
		testVector[i] = *bitvec.New(cellCount)
		vals := rand.Perm(cellCount)
		for j := 0; j < cellCount; j++ {
			if vals[j] > cellCount/2 {
				testVector[i].Set(j)
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.Read(testVector)
	}
}
