package pir

import (
	"math/rand"
	"testing"
)

const cellSize = 1024
const cellCount = 1024
const batchSize = 128

type boringDB struct {
	Data [][]byte
}

func (b boringDB) Read(n uint32) []byte {
	return b.Data[n]
}

func (b boringDB) Length() uint32 {
	return uint32(cellCount)
}

func BenchmarkRead64(b *testing.B) {
		DoRead(b, 64)
}

func BenchmarkRead128(b *testing.B) {
	DoRead(b, 128)
}

func BenchmarkRead256(b *testing.B) {
	DoRead(b, 256)
}

func BenchmarkRead512(b *testing.B) {
	DoRead(b, 512)
}


func DoRead(b *testing.B, cellMultiple int) {
	//Make Database
	var db boringDB
	theCellCount := cellCount * cellMultiple
	db.Data = make([][]byte, theCellCount)
	dataSize := cellSize * theCellCount
	fullData := make([]byte, dataSize)
	for i := 0; i < theCellCount; i++ {
		offset := i * cellSize
		db.Data[i] = fullData[offset:offset + cellSize]
	}
	server := PIRServer{db}

	//Make testVector
	testVector := make([]BitVec, batchSize)
	for i := 0; i < batchSize; i++ {
		testVector[i] = *NewBitVec(theCellCount)
		vals := rand.Perm(theCellCount)
		for j := 0; j < theCellCount; j++ {
			if vals[j] > theCellCount/2 {
				testVector[i].Set(j)
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.Read(testVector)
	}
}
