package pirtest

import (
	"math/rand"
	"testing"

	"github.com/privacylab/talek/pir"
	"github.com/privacylab/talek/pir/common"
)

// FatalInterface is a abstracts out calls to Fatal and Fatalf
type FatalInterface interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

const (
	// TestBatchSize is the batch size used in tests
	TestBatchSize = 3
	// TestNumMessages is the number of messages on the database used in tests
	TestNumMessages = 32
	// TestMessageSize is the message size used in tests
	TestMessageSize = 8
	// TestDepth is the bucket depth used in tests
	TestDepth = 2 // 16 buckets
	// BenchBatchSize is the batch size used in benchmarks
	BenchBatchSize = 128 // This seems to provide the best GPU perf
	// BenchNumMessages is the number of messages on the database used in benchmarks
	BenchNumMessages = 1048576 // 2^20
	// BenchMessageSize is the message size used in benchmarks
	BenchMessageSize = 1024
	// BenchDepth is the bucket depth used in benchmarks
	BenchDepth = 4 // 262144=2^18 buckets
//BenchNumMessages = 524288 // 2^19; Note: AMD devices have a smaller max memory allocation size
)

// AfterEach frees up the shard and context used in a test
func AfterEach(f FatalInterface, shard common.Shard, context pir.Context) {
	var err error
	if shard != nil {
		err = shard.Free()
		if err != nil {
			f.Fatalf("error freeing shard: %v\n", err)
		}
	}
	if context != nil {
		err = context.Free()
		if err != nil {
			f.Fatalf("error freeing context: %v\n", err)
		}
	}
}

// GenerateData will generate a byte array for tests
func GenerateData(size int) []byte {
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		data[i] = byte(i)
	}
	return data
}

// HelperTestShardRead is the generic function for testing correctness of a PIR implementation
func HelperTestShardRead(t *testing.T, shard common.Shard) {

	// Populate batch read request
	reqLength := shard.GetNumBuckets() / 8
	if shard.GetNumBuckets()%8 != 0 {
		reqLength++
	}
	reqs := make([]byte, reqLength*TestBatchSize)
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
	//fmt.Printf("%v\n", response)

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
			t.Fatalf("response2 is incorrect. byte %d was %d, not '%d'\n", i, res[i], expected)
		}
	}

	// Try to a malformed Read
	_, err = shard.Read(reqs, 7)
	if err == nil {
		t.Fatalf("shard.Read should have returned an error with mismatched reqs and reqLength")
	}
}

// HelperBenchmarkShardRead is the generic function for testing performance of a PIR implementation
func HelperBenchmarkShardRead(b *testing.B, shard common.Shard, batchSize int) {
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
}
