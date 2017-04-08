package server

import (
	"crypto/rand"
	"testing"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/libtalek"
	"github.com/privacylab/talek/pir"
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
	t0s := getSocket()
	t0c := make(chan int)
	go pir.CreateMockServer(t0c, t0s)
	<-t0c
	t0 := NewCentralized("t0", t0s, Config{&config, 1, 0, 0, nil, 0})

	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = t0.Write(repArgs, &reply)
	}

	t0c <- 1
}
