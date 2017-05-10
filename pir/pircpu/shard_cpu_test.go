package pircpu

import (
	"fmt"
	pt "github.com/privacylab/talek/pir/pirtest"
	"testing"
)

func TestShardCPUReadv0(t *testing.T) {
	fmt.Printf("TestShardCPUReadv0: ...\n")
	shard, err := NewShardCPU("shardcpuv0", pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), 0)
	if err != nil {
		t.Fatalf("cannot create new ShardCPU v0: error=%v\n", err)
	}
	pt.HelperTestShardRead(t, shard)
	pt.AfterEach(t, shard, nil)
	fmt.Printf("... done \n")
}

func TestShardCPUReadv1(t *testing.T) {
	fmt.Printf("TestShardCPUReadv1: ...\n")
	shard, err := NewShardCPU("shardcpuv1", pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), 1)
	if err != nil {
		t.Fatalf("cannot create new ShardCPU v1: error=%v\n", err)
	}
	pt.HelperTestShardRead(t, shard)
	pt.AfterEach(t, shard, nil)
	fmt.Printf("... done \n")
}

func BenchmarkShardCPUReadv0(b *testing.B) {
	shard, err := NewShardCPU("shardcpuv0", pt.BenchDepth*pt.BenchMessageSize, pt.GenerateData(pt.BenchNumMessages*pt.BenchMessageSize), 0)
	if err != nil {
		b.Fatalf("cannot create new ShardCPU v0: error=%v\n", err)
	}
	pt.HelperBenchmarkShardRead(b, shard, pt.BenchBatchSize)
	pt.AfterEach(b, shard, nil)
}

func BenchmarkShardCPUReadv1(b *testing.B) {
	shard, err := NewShardCPU("shardcpuv1", pt.BenchDepth*pt.BenchMessageSize, pt.GenerateData(pt.BenchNumMessages*pt.BenchMessageSize), 1)
	if err != nil {
		b.Fatalf("cannot create new ShardCPU v1: error=%v\n", err)
	}
	pt.HelperBenchmarkShardRead(b, shard, pt.BenchBatchSize)
	pt.AfterEach(b, shard, nil)
}
