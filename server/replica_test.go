package server

import (
	"crypto/rand"
	"testing"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/libtalek"
)

func BenchmarkWrite(b *testing.B) {
	config := common.Config{}
	config.NumBuckets = 128
	config.BucketDepth = 4
	config.DataSize = 256
	config.MaxLoadFactor = 0.90
	config.BloomFalsePositive = 0.1

	plaintext := make([]byte, config.DataSize, config.DataSize)
	_, err := rand.Read(plaintext)
	if err != nil {
		b.Fatalf("Error creating plaintext: %v\n", err)
	}
	th, err := libtalek.NewTopic()
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	args, err := th.GeneratePublish(&config, plaintext)
	if err != nil {
		b.Fatalf("Error creating WriteArgs: %v\n", err)
	}
	repArgs := &common.ReplicaWriteArgs{
		WriteArgs: *args,
		EpochFlag: false,
	}

	var reply common.ReplicaWriteReply
	t0 := NewReplica("t0", "cpu.0", Config{&config, 1, 0, 0, nil, 0})

	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = t0.Write(repArgs, &reply)
	}

}
