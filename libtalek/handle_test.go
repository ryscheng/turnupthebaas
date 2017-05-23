package libtalek

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/privacylab/talek/common"
)

func TestGeneratePoll(t *testing.T) {
	fmt.Printf("TestGeneratePoll:\n")
	config := &ClientConfig{&common.Config{}, 0, 0, nil, ""}
	config.Config.NumBuckets = 1000000
	config.TrustDomains = make([]*common.TrustDomainConfig, 3)

	h, err := NewHandle()
	if err != nil {
		t.Fatalf("Error creating handle: %v\n", err)
	}
	_, _, err = h.generatePoll(config, rand.Reader)
	if err == nil {
		t.Fatalf("Could generate a poll from an un-configured subscription")
	}

	topic, _ := NewTopic()
	h = &topic.Handle
	args0, _, err := h.generatePoll(config, rand.Reader)
	if err != nil {
		t.Fatalf("Error creating ReadArgs: %v\n", err)
	}

	if uint64(len(args0.TD[0].RequestVector)) != config.Config.NumBuckets/8 {
		t.Fatalf("Length of request was incorrect. %d vs %d", len(args0.TD[0].RequestVector), config.Config.NumBuckets/8)
	}

	fmt.Printf("len(args0)=%v; \n", 3*(len(args0.TD[0].RequestVector)+len(args0.TD[0].PadSeed)))

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
	config := &ClientConfig{&common.Config{}, 0, 0, nil, ""}
	config.TrustDomains = make([]*common.TrustDomainConfig, 3)
	config.Config.NumBuckets = NumBuckets

	topic, err := NewTopic()
	h := topic.Handle
	if err != nil {
		b.Fatalf("Error creating handle: %v\n", err)
	}
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = h.generatePoll(config, rand.Reader)
	}

}

func BenchmarkRetrieveResponse(b *testing.B) {
	config := &ClientConfig{&common.Config{}, 0, 0, nil, ""}
	config.TrustDomains = make([]*common.TrustDomainConfig, 3)
	config.Config.NumBuckets = 10

	topic, err := NewTopic()
	h := topic.Handle
	if err != nil {
		b.Fatalf("Error creating topic handle: %v\n", err)
	}
	args, _, err := h.generatePoll(config, rand.Reader)
	if err != nil {
		b.Fatalf("Error creating ReadArgs: %v\n", err)
	}
	reply := &common.ReadReply{}
	reply.Data = make([]byte, 1024)
	// Start timing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.retrieveResponse(args, reply, 1024)
	}

}
