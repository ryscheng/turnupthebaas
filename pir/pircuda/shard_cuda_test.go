//+build !nocuda,!travis

package pircuda

import (
	"fmt"
	"testing"

	pt "github.com/privacylab/talek/pir/pirtest"
)

func TestShardCUDAReadv0(t *testing.T) {
	fmt.Printf("TestShardCUDAReadv0: ...\n")
	context, err := NewContextCUDA("contextcuda", "/kernel.ptx", 8)
	if err != nil {
		t.Fatalf("cannot create new ContextCUDA: error=%v\n", err)
	}
	shard, err := NewShardCUDA("shardcuda", context, pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), pt.TestNumMessages*pt.TestMessageSize/context.GetKernelDataSize())
	if err != nil {
		t.Fatalf("cannot create new ShardCUDA: error=%v\n", err)
	}
	pt.HelperTestShardRead(t, shard)
	pt.AfterEach(t, shard, context)
	fmt.Printf("... done \n")
}

func BenchmarkShardCUDAReadv0(b *testing.B) {
	context, err := NewContextCUDA("contextcuda", "/kernel.ptx", 8)
	if err != nil {
		b.Fatalf("cannot create new ShardCUDA: error=%v\n", err)
	}
	shard, err := NewShardCUDA("shardcuda", context, pt.BenchDepth*pt.BenchMessageSize, pt.GenerateData(pt.BenchNumMessages*pt.BenchMessageSize), pt.BenchNumMessages*pt.BenchMessageSize/context.GetKernelDataSize())
	if err != nil {
		b.Fatalf("cannot create new ShardCUDA: error=%v\n", err)
	}
	pt.HelperBenchmarkShardRead(b, shard, pt.BenchBatchSize)
	pt.AfterEach(b, shard, context)
}
