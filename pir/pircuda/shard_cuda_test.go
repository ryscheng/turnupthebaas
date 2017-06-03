//+build !nocuda,!travis

package pircuda

import (
	"fmt"
	"testing"

	"github.com/privacylab/talek/common"
	pt "github.com/privacylab/talek/pir/pirtest"
)

func beforeEach() {
	common.SilenceLoggers()
}

func TestShardCUDACreate(t *testing.T) {
	fmt.Printf("TestShardCUDACreate: ...\n")
	beforeEach()
	context, err := NewContextCUDA("contextcuda", "/kernel.ptx", 8)
	if err != nil {
		t.Fatalf("cannot create new ContextCUDA: error=%v\n", err)
	}
	shard, err := NewShardCUDA("shardcuda", context, 7, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), pt.TestNumMessages*pt.TestMessageSize/context.GetKernelDataSize())
	if err == nil {
		t.Fatalf("new ShardCUDA should have failed with invalid bucketSize, but didn't return error")
	}
	if shard != nil {
		t.Fatalf("new ShardCUDA should have failed with invalid bucketSize, but returned a shard")
	}
	pt.AfterEach(t, nil, context)
	fmt.Printf("... done \n")
}

func TestShardCUDAReadv0(t *testing.T) {
	fmt.Printf("TestShardCUDAReadv0: ...\n")
	beforeEach()
	context, err := NewContextCUDA("contextcuda", "/kernel.ptx", 8)
	if err != nil {
		t.Fatalf("cannot create new ContextCUDA: error=%v\n", err)
	}
	shard, err := NewShardCUDA("shardcuda", context, pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), pt.TestNumMessages*pt.TestMessageSize/context.GetKernelDataSize())
	if err != nil {
		t.Fatalf("cannot create new ShardCUDA: error=%v\n", err)
	}
	pt.HelperTestShardRead(t, shard)
	pt.HelperTestClientRead(t, shard)
	pt.AfterEach(t, shard, context)
	fmt.Printf("... done \n")
}

func BenchmarkShardCUDAReadv0(b *testing.B) {
	beforeEach()
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
