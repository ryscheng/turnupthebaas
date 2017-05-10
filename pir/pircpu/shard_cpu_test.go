package pir

import (
	"fmt"
	"testing"
)

func TestShardCPUReadv0(t *testing.T) {
	fmt.Printf("TestShardCPUReadv0: ...\n")
	shard, err := NewShardCPU("shardcpuv0", TestDepth*TestMessageSize, generateData(TestNumMessages*TestMessageSize), 0)
	if err != nil {
		t.Fatalf("cannot create new ShardCPU v0: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
	afterEach(t, shard, nil)
	fmt.Printf("... done \n")
}

func TestShardCPUReadv1(t *testing.T) {
	fmt.Printf("TestShardCPUReadv1: ...\n")
	shard, err := NewShardCPU("shardcpuv1", TestDepth*TestMessageSize, generateData(TestNumMessages*TestMessageSize), 1)
	if err != nil {
		t.Fatalf("cannot create new ShardCPU v1: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
	afterEach(t, shard, nil)
	fmt.Printf("... done \n")
}

func BenchmarkShardCPUReadv0(b *testing.B) {
	shard, err := NewShardCPU("shardcpuv0", BenchDepth*BenchMessageSize, generateData(BenchNumMessages*BenchMessageSize), 0)
	if err != nil {
		b.Fatalf("cannot create new ShardCPU v0: error=%v\n", err)
	}
	HelperBenchmarkShardRead(b, shard, BenchBatchSize)
	afterEach(b, shard, nil)
}

func BenchmarkShardCPUReadv1(b *testing.B) {
	shard, err := NewShardCPU("shardcpuv1", BenchDepth*BenchMessageSize, generateData(BenchNumMessages*BenchMessageSize), 1)
	if err != nil {
		b.Fatalf("cannot create new ShardCPU v1: error=%v\n", err)
	}
	HelperBenchmarkShardRead(b, shard, BenchBatchSize)
	afterEach(b, shard, nil)
}
