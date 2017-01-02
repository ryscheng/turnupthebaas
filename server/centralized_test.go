package server

import (
	"crypto/rand"
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/libpdb"
	"github.com/ryscheng/pdb/pir"
	"testing"
	"time"
)

func BenchmarkWrite(b *testing.B) {
	trustDomainConfig0 := common.NewTrustDomainConfig("t0", "localhost:9000", true, false)
	trustDomainConfig1 := common.NewTrustDomainConfig("t1", "localhost:9100", true, false)
	config := common.CommonConfig{0, 0, 0, 0, 0, 0, 0, time.Second, time.Second, []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1}}
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
	password := ""
	th, err := libpdb.NewTopic(password)
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	args, err := th.GeneratePublish(&config, 1, plaintext)
	if err != nil {
		b.Fatalf("Error creating WriteArgs: %v\n", err)
	}

	var reply common.WriteReply
	t1s := getSocket()
	t1c := make(chan int)
	go pir.CreateMockServer(t1c, t1s)
	<-t1c
	t1 := NewCentralized("t1", t1s, config, nil, false)

	t0s := getSocket()
	t0c := make(chan int)
	go pir.CreateMockServer(t0c, t0s)
	<-t0c
	t0 := NewCentralized("t0", t0s, config, t1, true)

	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = t0.Write(args, &reply)
	}

	t1c <- 1
	t0c <- 1
}
