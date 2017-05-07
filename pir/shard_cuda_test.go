//+build !travis

package pir

import (
	"fmt"
	"testing"
)

func TestShardCUDAReadv0(t *testing.T) {
	fmt.Printf("TestShardCUDAReadv0: ...\n")
	context, err := NewContextCUDA("contextcuda", "/pir/cuda_modules/pir.ptx")
	if err != nil {
		t.Fatalf("cannot create new ContextCUDA: error=%v\n", err)
	}
	shard, err := NewShardCUDA("shardcuda", context, TestDepth*TestMessageSize, generateData(TestNumMessages*TestMessageSize), TestBatchSize*context.GetGroupSize())
	if err != nil {
		t.Fatalf("cannot create new ShardCUDA: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
	afterEach(t, shard, context)
	fmt.Printf("... done \n")
}

func BenchmarkShardCUDAReadv0(b *testing.B) {
	batchSize := 1
	context, err := NewContextCUDA("contextcuda", "/pir/cuda_modules/pir.ptx")
	if err != nil {
		b.Fatalf("cannot create new ShardCUDA: error=%v\n", err)
	}
	shard, err := NewShardCUDA("shardcuda", context, BenchDepth*BenchMessageSize, generateData(BenchNumMessages*BenchMessageSize), batchSize*context.GetGroupSize())
	if err != nil {
		b.Fatalf("cannot create new ShardCUDA: error=%v\n", err)
	}
	HelperBenchmarkShardRead(b, shard, batchSize)
	afterEach(b, shard, context)
}
