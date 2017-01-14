package libpdb

import (
	"fmt"
	"github.com/ryscheng/pdb/common"
	"testing"
)

func TestGeneratePoll(t *testing.T) {
	fmt.Printf("TestGeneratePoll:\n")
	config := &ClientConfig{&common.CommonConfig{}, 0, 0, nil}
	config.CommonConfig.NumBuckets = 1000000
	config.TrustDomains = make([]*common.TrustDomainConfig, 3)

	password := ""
	sub, err := NewSubscription(0)
	if err != nil {
		t.Fatalf("Error creating subscription handle: %v\n", err)
	}
	args0, _, err := sub.generatePoll(config, 1)
	if err != nil {
		t.Fatalf("Error creating ReadArgs: %v\n", err)
	}

	fmt.Printf("len(args0)=%v; \n", 3*(len(args0.ForTd[0].RequestVector)+len(args0.ForTd[0].PadSeed)))

	fmt.Printf("... done \n")
}


func BenchmarkGeneratePollN10K(b *testing.B) {
	HelperBenchmarkGeneratePoll(b, 10000/4)
}
func BenchmarkGeneratePollN100K(b *testing.B) {
	HelperBenchmarkGeneratePoll(b, 100000/4)
}
func BenchmarkGeneratePollN1M(b *testing.B) {
	HelperBenchmarkGeneratePoll(b, 1000000/4)
}

func HelperBenchmarkGeneratePoll(b *testing.B, NumBuckets uint64) {
	config := &ClientConfig{&common.CommonConfig{}, 0, 0, nil}
	config.TrustDomains = make([]*common.TrustDomainConfig, 3)
	config.CommonConfig.NumBuckets = NumBuckets

	password := ""
	sub, err := NewSubscription(0)
	if err != nil {
		b.Fatalf("Error creating subscription handle: %v\n", err)
	}
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = sub.generatePoll(config, uint64(i))
	}

}

func BenchmarkRetrieveResponse(b *testing.B) {
	config := &ClientConfig{&common.CommonConfig{}, 0, 0, nil}
	config.TrustDomains = make([]*common.TrustDomainConfig, 3)
	config.CommonConfig.NumBuckets = 10

	password := ""
	sub, err := NewSubscription(0)
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	args, _, err := sub.generatePoll(config, 0)
	if err != nil {
		b.Fatalf("Error creating ReadArgs: %v\n", err)
	}
	reply := &common.ReadReply{}
	reply.Data = make([]byte, 1024)
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sub.retrieveResponse(args, reply)
	}

}
