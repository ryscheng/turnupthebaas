package pir

import (
	"fmt"
	"math/rand"
	"testing"
)

func generateData(size int) []byte {
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		data[i] = byte(i)
	}
	return data
}

func HelperTestShardRead(t *testing.T, shard Shard) {
	fmt.Printf("TestShardRead: %s ...\n", shard.GetName())

	// Populate batch read request
	batchSize := 3
	reqLength := shard.GetNumBuckets() / 8
	if shard.GetNumBuckets()%8 != 0 {
		reqLength++
	}
	reqs := make([]byte, reqLength*batchSize)
	setBit := func(reqs []byte, reqIndex int, bucketIndex int) {
		reqs[reqIndex*reqLength+(bucketIndex/8)] |= byte(1) << uint(bucketIndex%8)
	}
	setBit(reqs, 0, 1)
	setBit(reqs, 1, 0)
	setBit(reqs, 2, 0)
	setBit(reqs, 2, 1)
	setBit(reqs, 2, 2)

	if shard.GetNumBuckets() < 3 {
		t.Fatalf("test misconfigured. shard has %d buckets, needs %d\n", shard.GetNumBuckets(), 3)
	}

	// Batch Read
	response, err := shard.Read(reqs, reqLength)

	// Check fail
	if err != nil {
		t.Fatalf("error calling shard.Read: %v\n", err)
	}

	if response == nil {
		t.Fatalf("no response received")
	}

	bucketSize := shard.GetBucketSize()
	data := shard.GetData()
	// Check request 0
	res := response[0:bucketSize]
	for i := 0; i < bucketSize; i++ {
		if res[i] != data[bucketSize+i] {
			t.Fatalf("response0 is incorrect. byte %d was %d, not '%d'\n", i, res[i], bucketSize+i)
		}
	}
	// Check request 1
	res = response[bucketSize : 2*bucketSize]
	for i := 0; i < bucketSize; i++ {
		if res[i] != data[i] {
			t.Fatalf("response1 is incorrect. byte %d was %d, not '%d'\n", i, res[i], i)
		}
	}
	// Check request 2
	res = response[2*bucketSize : 3*bucketSize]
	for i := 0; i < bucketSize; i++ {
		expected := data[i] ^ data[bucketSize+i] ^ data[2*bucketSize+i]
		if res[i] != expected {
			t.Fatalf("response is incorrect. byte %d was %d, not '%d'\n", i, res[i], expected)
		}
	}

	// Free
	err = shard.Free()
	if err != nil {
		t.Fatalf("error freeing shard: %v\n", err)
	}

	fmt.Printf("... done \n")

}

func TestShardCPUReadv0(t *testing.T) {
	numMessages := 32
	messageSize := 2
	depth := 2 // 16 buckets
	shard, err := NewShardCPU("shardcpuv0", depth*messageSize, generateData(numMessages*messageSize), 0)
	if err != nil {
		t.Fatalf("cannot create new ShardCPU v0: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
}

func TestShardCPUReadv1(t *testing.T) {
	numMessages := 32
	messageSize := 2
	depth := 2 // 16 buckets
	shard, err := NewShardCPU("shardcpuv1", depth*messageSize, generateData(numMessages*messageSize), 1)
	if err != nil {
		t.Fatalf("cannot create new ShardCPU v1: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
}

func XTestShardCLRead(t *testing.T) {
	numMessages := 32
	messageSize := 2
	depth := 2 // 16 buckets
	shard, err := NewShardCL("shardcl", depth*messageSize, generateData(numMessages*messageSize), 0)
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
}

func HelperBenchmarkShardRead(b *testing.B, shard Shard, batchSize int) {
	reqLength := shard.GetNumBuckets() / 8
	if shard.GetNumBuckets()%8 != 0 {
		reqLength++
	}
	reqs := make([]byte, reqLength*batchSize)
	for i := 0; i < len(reqs); i++ {
		reqs[i] = byte(rand.Int())
	}

	// Start test
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := shard.Read(reqs, reqLength)

		if err != nil {
			b.Fatalf("Read error: %v\n", err)
		}
	}
	b.StopTimer()
	// Free
	err := shard.Free()
	if err != nil {
		b.Fatalf("error freeing shard: %v\n", err)
	}

}

func BenchmarkShardCPUReadv0(b *testing.B) {
	numMessages := 1000000
	messageSize := 1024
	depth := 4 // 250000 buckets
	shard, err := NewShardCPU("shardcpuv0", depth*messageSize, generateData(numMessages*messageSize), 0)
	if err != nil {
		b.Fatalf("cannot create new ShardCPU v0: error=%v\n", err)
	}
	HelperBenchmarkShardRead(b, shard, 8)
}

func BenchmarkShardCPUReadv1(b *testing.B) {
	numMessages := 1000000
	messageSize := 1024
	depth := 4 // 250000 buckets
	shard, err := NewShardCPU("shardcpuv1", depth*messageSize, generateData(numMessages*messageSize), 1)
	if err != nil {
		b.Fatalf("cannot create new ShardCPU v1: error=%v\n", err)
	}
	HelperBenchmarkShardRead(b, shard, 8)
}

func XBenchmarkShardCLRead(b *testing.B) {
	numMessages := 1000000
	messageSize := 1024
	depth := 4 // 250000 buckets
	shard, err := NewShardCL("shardcl", depth*messageSize, generateData(numMessages*messageSize), 0)
	if err != nil {
		b.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperBenchmarkShardRead(b, shard, 8)
}
