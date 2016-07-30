package pir

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

const cellSize = 1024
const cellCount = 1024
const batchSize = 512

type boringDB struct {
	DataLength uint32
	Data [][]byte
}

func (b boringDB) Read(n uint32) []byte {
	return b.Data[n]
}

func (b boringDB) Length() uint32 {
	return b.DataLength
}

func TestSpeeds(t *testing.T) {
	var counts = []int{1024, 1024 * 4, 1024 * 16, 1024 * 64, 1024 * 256}
	var sizes = []int{1024, 2048, 4096}
	var chunks = []int{16, 32, 64, 128, 256}
	fmt.Printf("number of cells, size of cell, batched requests, nanoseconds\n")
	for _, count := range counts {
		for _, size := range sizes {
			for _, chunk := range chunks {
				timing := SpeedFor(count, size, chunk)
				fmt.Printf("%d,%d,%d,%d\n", count, size, chunk, timing.Nanoseconds())
			}
		}
	}
}

func SpeedFor(count int, size int, chunk int) time.Duration {
	//Make Database
	var db boringDB
  db.DataLength = uint32(count)
	db.Data = make([][]byte, count)
	dataSize := count * size
	fullData := make([]byte, dataSize)
	for i := 0; i < dataSize; i++ {
		fullData[i] = 1
	}
	for i := 0; i < count; i++ {
		offset := i * size
		db.Data[i] = fullData[offset:offset + size]
	}
	server := PIRServer{db}

	//Make testVector
	testVector := make([]BitVec, chunk)
	for i := 0; i < chunk; i++ {
		testVector[i] = *NewBitVec(count)
		vals := rand.Perm(count)
		for j := 0; j < count; j++ {
			if vals[j] > count/2 {
				testVector[i].Set(j)
			}
		}
	}

	then := time.Now()
	server.Read(testVector)
	return time.Since(then)
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
	db.DataLength = uint32(theCellCount)
	db.Data = make([][]byte, theCellCount)
	dataSize := cellSize * theCellCount
	fullData := make([]byte, dataSize)
	for i := 0; i < dataSize; i++ {
		fullData[i] = byte(rand.Int63())
	}
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
