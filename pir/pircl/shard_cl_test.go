//+build !noopencl,!travis

package pircl

import (
	"fmt"
	"testing"

	pt "github.com/privacylab/talek/pir/pirtest"
)

func TestShardCLReadv0(t *testing.T) {
	fmt.Printf("TestShardCLReadv0: ...\n")
	context, err := NewContextCL("contextcl", KernelCL0, 8, pt.BenchMessageSize*pt.BenchDepth)
	if err != nil {
		t.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), pt.TestBatchSize*context.GetGroupSize())
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	pt.HelperTestShardRead(t, shard)
	pt.AfterEach(t, shard, context)
	fmt.Printf("... done \n")
}

func TestShardCLReadv1(t *testing.T) {
	fmt.Printf("TestShardCLReadv1: ...\n")
	context, err := NewContextCL("contextcl", KernelCL1, 8, 1)
	if err != nil {
		t.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), pt.TestBatchSize*pt.TestDepth*pt.TestMessageSize/context.GetKernelDataSize())
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	pt.HelperTestShardRead(t, shard)
	pt.AfterEach(t, shard, context)
	fmt.Printf("... done \n")
}

func TestShardCLReadv2(t *testing.T) {
	fmt.Printf("TestShardCLReadv2: ...\n")
	context, err := NewContextCL("contextcl", KernelCL2, 8, 1)
	if err != nil {
		t.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), pt.TestNumMessages*pt.TestMessageSize/context.GetKernelDataSize())
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	pt.HelperTestShardRead(t, shard)
	pt.AfterEach(t, shard, context)
	fmt.Printf("... done \n")
}

func BenchmarkShardCLReadv0(b *testing.B) {
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
