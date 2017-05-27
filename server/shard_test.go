package server

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/privacylab/talek/common"
	_ "github.com/privacylab/talek/pir/pircpu"
)

import "testing"

func fromEnvOrDefault(envKey string, defaultVal int) int {
	if os.Getenv(envKey) != "" {
		val, _ := strconv.Atoi(os.Getenv(envKey))
		return val
	}
	return defaultVal
}

func testConf() Config {
	return Config{
		Config: &common.Config{
			NumBuckets:         uint64(fromEnvOrDefault("NUM_BUCKETS", 512)),
			BucketDepth:        uint64(fromEnvOrDefault("BUCKET_DEPTH", 4)),
			DataSize:           uint64(fromEnvOrDefault("DATA_SIZE", 512)),
			BloomFalsePositive: 0.95,
			MaxLoadFactor:      0.95,
			LoadFactorStep:     0.02,
		},
		ReadBatch:        fromEnvOrDefault("BATCH_SIZE", 8),
		WriteInterval:    time.Second,
		ReadInterval:     time.Second,
		TrustDomain:      nil,
		TrustDomainIndex: 0,
	}
}

func TestShardSanity(t *testing.T) {
	conf := testConf()
	shard := NewShard("Test Shard", "cpu.0", conf)
	if shard == nil {
		t.Error("Failed to create shard.")
		return
	}

	writeReplyChan := make(chan *common.WriteReply)

	data := make([]byte, conf.Config.DataSize)
	copy(data, bytes.NewBufferString("Magic").Bytes())
	shard.Write(&common.ReplicaWriteArgs{
		WriteArgs: common.WriteArgs{
			Bucket1:        0,
			Bucket2:        1,
			Data:           data,
			InterestVector: []uint64{},
			ReplyChan:      writeReplyChan,
		},
		EpochFlag: false,
	})

	// Force DB write.
	shard.syncChan <- 1

	replychan := make(chan *common.BatchReadReply)

	rv := make([]byte, 512)
	rv[0] = 0xff
	req := common.PirArgs{RequestVector: rv}
	reqs := make([]common.PirArgs, 8)
	for i := 0; i < 8; i++ {
		reqs[i] = req
	}
	shard.BatchRead(&DecodedBatchReadRequest{Args: reqs, ReplyChan: replychan})

	reply := <-replychan
	if reply.Replies[0].Data[0] != bytes.NewBufferString("Magic").Bytes()[0] {
		t.Error("Failed to round-trip a write.")
		return
	}

	shard.Close()
}

func BenchmarkShard(b *testing.B) {
	fmt.Printf("Benchmark began with N=%d\n", b.N)
	readsPerWrite := fromEnvOrDefault("READS_PER_WRITE", 20)

	conf := testConf()
	shard := NewShard("Test Shard", "cpu.0", conf)
	if shard == nil {
		b.Error("Failed to create shard.")
		return
	}

	replychan := make(chan *common.BatchReadReply)

	//A default write request
	stdWrite := common.WriteArgs{
		Bucket1:        0,
		Bucket2:        1,
		Data:           bytes.NewBufferString("Magic").Bytes(),
		InterestVector: []uint64{},
	}
	shardWrite := &common.ReplicaWriteArgs{
		WriteArgs: stdWrite,
		EpochFlag: false,
	}

	//A default read request
	reqs := make([]common.PirArgs, conf.ReadBatch)
	rv := make([]byte, int(conf.NumBuckets))
	for i := 0; i < len(rv); i++ {
		rv[i] = byte(rand.Int())
	}
	req := common.PirArgs{RequestVector: rv}
	for i := 0; i < conf.ReadBatch; i++ {
		reqs[i] = req
	}
	stdRead := &DecodedBatchReadRequest{reqs, replychan}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if i%readsPerWrite == 0 {
			stdWrite.Bucket1 = uint64(rand.Int()) % conf.NumBuckets
			stdWrite.Bucket2 = uint64(rand.Int()) % conf.NumBuckets
			shard.Write(shardWrite)
		} else {
			shard.BatchRead(stdRead)
			reply := <-replychan

			if reply == nil || reply.Err != "" {
				b.Error("Read failed.")
			}
		}
		b.SetBytes(int64(1))
	}

	fmt.Printf("Benchmark called close w N=%d\n", b.N)
	shard.Close()
}
