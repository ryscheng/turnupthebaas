//+build !travis

package pir

import (
	"testing"
)

func TestShardCLRead(t *testing.T) {
	numMessages := 32
	messageSize := 2
	depth := 2 // 16 buckets
	shard, err := NewShardCL("shardcl", depth*messageSize, generateData(numMessages*messageSize), 0)
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
}

func BenchmarkShardCLRead(b *testing.B) {
	numMessages := 1000000
	messageSize := 1024
	depth := 4 // 250000 buckets
	shard, err := NewShardCL("shardcl", depth*messageSize, generateData(numMessages*messageSize), 0)
	if err != nil {
		b.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperBenchmarkShardRead(b, shard, 8)
}
