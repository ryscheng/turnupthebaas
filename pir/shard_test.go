package pir

import (
	"fmt"
	"testing"

	"github.com/willf/bitset"
)

func generateData(size int) []byte {
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		data[i] = byte(i)
	}
	return data
}

func HelperTestShardRead(t *testing.T, shard Shard) {
	fmt.Printf("TestShardCPURead: %s ...\n", shard.GetName())

	reqs := make([]bitset.BitSet, 3)
	reqs[0].SetTo(1, true)
	reqs[1].SetTo(0, true)
	reqs[2].SetTo(0, true)
	reqs[2].SetTo(1, true)
	reqs[2].SetTo(2, true)

	if shard.GetNumBuckets() < 3 {
		t.Fatalf("test misconfigured. shard has %d buckets, needs %d\n", shard.GetNumBuckets(), 3)
	}

	response, err := shard.Read(reqs)

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

	err = shard.Free()
	if err != nil {
		t.Fatalf("error freeing shard: %v\n", err)
	}

	fmt.Printf("... done \n")

}

func TestShardCPURead(t *testing.T) {
	numMessages := 32
	messageSize := 2
	depth := 2 // 16 buckets
	shard, err := NewShardCPU("shardcpuv0", depth*messageSize, generateData(numMessages*messageSize), 0)
	if err != nil {
		t.Fatalf("cannot create new ShardCPU v0: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
	shard, err = NewShardCPU("shardcpuv1", depth*messageSize, generateData(numMessages*messageSize), 1)
	if err != nil {
		t.Fatalf("cannot create new ShardCPU v1: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
}

/**
func BenchmarkPir(b *testing.B) {
	cellLength := 1024
	cellCount := 2048
	batchSize := 8
	if os.Getenv("PIR_CELL_LENGTH") != "" {
		cellLength, _ = strconv.Atoi(os.Getenv("PIR_CELL_LENGTH"))
	}
	if os.Getenv("PIR_CELL_COUNT") != "" {
		cellCount, _ = strconv.Atoi(os.Getenv("PIR_CELL_COUNT"))
	}
	if os.Getenv("PIR_BATCH_SIZE") != "" {
		batchSize, _ = strconv.Atoi(os.Getenv("PIR_BATCH_SIZE"))
	}

	sockName := getSocket()
	status := make(chan int)
	go CreateMockServer(status, sockName)
	<-status

	pirServer, err := Connect(sockName)
	if err != nil {
		b.Error(err)
		return
	}

	pirServer.Configure(cellLength, cellCount, batchSize)
	db, err := pirServer.GetDB()
	if err != nil {
		b.Error(err)
		return
	}
	for x := range db.DB {
		db.DB[x] = byte(x)
	}

	pirServer.SetDB(db)

	responseChan := make(chan []byte)
	masks := make([]byte, cellCount*batchSize/8)
	for i := 0; i < len(masks); i++ {
		masks[i] = byte(rand.Int())
	}

	b.ResetTimer()

	signalChan := make(chan int)
	go func() {
		for j := 0; j < b.N; j++ {
			response := <-responseChan
			b.SetBytes(int64(len(response)))
		}
		signalChan <- 1
	}()

	for i := 0; i < b.N; i++ {
		err := pirServer.Read(masks, responseChan)

		if err != nil {
			b.Error(err)
		}
	}

	<-signalChan

	pirServer.Disconnect()

	status <- 1
	<-status
}
**/
