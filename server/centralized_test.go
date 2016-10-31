package server

import (
	"crypto/rand"
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/libpdb"
	"testing"
	"time"
)

func BenchmarkWrite(b *testing.B) {
	trustDomainConfig0 := common.NewTrustDomainConfig("t0", "localhost:9000", true, false)
	trustDomainConfig1 := common.NewTrustDomainConfig("t1", "localhost:9100", true, false)
	globalConfig := common.GlobalConfig{0, 0, 0, 0, 0, 0, time.Second, time.Second, []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1}}
	globalConfig.NumBuckets = 100
	globalConfig.BucketDepth = 4
	globalConfig.WindowSize = 100000
	globalConfig.DataSize = 256
	globalConfig.BloomFalsePositive = 0.1

	plaintext := make([]byte, globalConfig.DataSize, globalConfig.DataSize)
	_, err := rand.Read(plaintext)
	if err != nil {
		b.Fatalf("Error creating plaintext: %v\n", err)
	}
	password := ""
	th, err := libpdb.NewTopicHandle(password)
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	args, err := th.Publish(&globalConfig, 1, plaintext)
	if err != nil {
		b.Fatalf("Error creating WriteArgs: %v\n", err)
	}

	var reply common.WriteReply
	t1 := NewCentralized("t1", globalConfig, nil, false)
	t0 := NewCentralized("t0", globalConfig, t1, true)

	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = t0.Write(args, &reply)
	}

}
