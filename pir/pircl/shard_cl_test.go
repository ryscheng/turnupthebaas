//+build !noopencl,!travis

package pircl

import (
	"fmt"
	"testing"

	"github.com/privacylab/talek/common"
	pt "github.com/privacylab/talek/pir/pirtest"
)

func beforeEach() {
	common.SilenceLoggers()
}

func TestShardCLCreate(t *testing.T) {
	fmt.Printf("TestShardCLCreate: ...\n")
	beforeEach()
	context, err := NewContextCL("contextcl", KernelCL0, 8, pt.BenchMessageSize*pt.BenchDepth)
	if err != nil {
		t.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	// Creating with invalid bucketSize
	shard, err := NewShardCL("shardcl", context, 7, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), pt.TestBatchSize*context.GetGroupSize())
	if err == nil {
		t.Fatalf("new ShardCL should have failed with invalid bucketSize, but didn't return error")
	}
	if shard != nil {
		t.Fatalf("new ShardCL should have failed with invalid bucketSize, but returned a shard")
	}
	pt.AfterEach(t, nil, context)
	fmt.Printf("... done \n")

}

func TestShardCLReadv0(t *testing.T) {
	fmt.Printf("TestShardCLReadv0: ...\n")
	beforeEach()
	context, err := NewContextCL("contextcl", KernelCL0, 8, pt.BenchMessageSize*pt.BenchDepth)
	if err != nil {
		t.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), pt.TestBatchSize*context.GetGroupSize())
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	pt.HelperTestShardRead(t, shard)
	pt.HelperTestClientRead(t, shard)
	pt.AfterEach(t, shard, context)
	fmt.Printf("... done \n")
}

func TestShardCLReadv1(t *testing.T) {
	fmt.Printf("TestShardCLReadv1: ...\n")
	beforeEach()
	context, err := NewContextCL("contextcl", KernelCL1, 8, 1)
	if err != nil {
		t.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), pt.TestBatchSize*pt.TestDepth*pt.TestMessageSize/context.GetKernelDataSize())
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	pt.HelperTestShardRead(t, shard)
	pt.HelperTestClientRead(t, shard)
	pt.AfterEach(t, shard, context)
	fmt.Printf("... done \n")
}

func TestShardCLReadv2(t *testing.T) {
	fmt.Printf("TestShardCLReadv2: ...\n")
	beforeEach()
	context, err := NewContextCL("contextcl", KernelCL2, 8, 1)
	if err != nil {
		t.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), pt.TestNumMessages*pt.TestMessageSize/context.GetKernelDataSize())
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	pt.HelperTestShardRead(t, shard)
	pt.HelperTestClientRead(t, shard)
	pt.AfterEach(t, shard, context)
	fmt.Printf("... done \n")
}

func BenchmarkShardCLReadv0(b *testing.B) {
	//fmt.Printf("BenchmarkShardCLReadv0 began with N=%d... \n", b.N)
	beforeEach()
	context, err := NewContextCL("contextcl", KernelCL0, 8, pt.BenchMessageSize*pt.BenchDepth)
	if err != nil {
		b.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, pt.BenchDepth*pt.BenchMessageSize, pt.GenerateData(pt.BenchNumMessages*pt.BenchMessageSize), pt.BenchBatchSize*context.GetGroupSize())
	if err != nil {
		b.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	pt.HelperBenchmarkShardRead(b, shard, pt.BenchBatchSize)
	pt.AfterEach(b, shard, context)
}

func BenchmarkShardCLReadv1(b *testing.B) {
	//fmt.Printf("BenchmarkShardCLReadv1 began with N=%d... \n", b.N)
	beforeEach()
	context, err := NewContextCL("contextcl", KernelCL1, 8, 1)
	if err != nil {
		b.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, pt.BenchDepth*pt.BenchMessageSize, pt.GenerateData(pt.BenchNumMessages*pt.BenchMessageSize), pt.BenchBatchSize*pt.BenchDepth*pt.BenchMessageSize/context.GetKernelDataSize())
	if err != nil {
		b.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	pt.HelperBenchmarkShardRead(b, shard, pt.BenchBatchSize)
	pt.AfterEach(b, shard, context)
}

func BenchmarkShardCLReadv2(b *testing.B) {
	//fmt.Printf("BenchmarkShardCLReadv2 began with N=%d... \n", b.N)
	beforeEach()
	context, err := NewContextCL("contextcl", KernelCL2, 8, 1)
	if err != nil {
		b.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, pt.BenchDepth*pt.BenchMessageSize, pt.GenerateData(pt.BenchNumMessages*pt.BenchMessageSize), pt.BenchNumMessages*pt.BenchMessageSize/context.GetKernelDataSize())

	if err != nil {
		b.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	pt.HelperBenchmarkShardRead(b, shard, pt.BenchBatchSize)
	pt.AfterEach(b, shard, context)
}
