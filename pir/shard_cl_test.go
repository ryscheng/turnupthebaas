//+build !travis

package pir

import (
	"fmt"
	"testing"
)

func afterEachShardCL(f FatalInterface, shard Shard, context *ContextCL) {
	err := shard.Free()
	if err != nil {
		f.Fatalf("error freeing shard: %v\n", err)
	}
	err = context.Free()
	if err != nil {
		f.Fatalf("error freeing context: %v\n", err)
	}
}

func XTestShardCLReadv0(t *testing.T) {
	fmt.Printf("TestShardCLReadv1: ...\n")
	context, err := NewContextCL("contextcl", KernelCL0)
	if err != nil {
		t.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, TestDepth*TestMessageSize, generateData(TestNumMessages*TestMessageSize), TestBatchSize*context.GetGroupSize())
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
	afterEachShardCL(t, shard, context)
	fmt.Printf("... done \n")
}

func TestShardCLReadv1(t *testing.T) {
	fmt.Printf("TestShardCLReadv1: ...\n")
	context, err := NewContextCL("contextcl", KernelCL1)
	if err != nil {
		t.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, TestDepth*TestMessageSize, generateData(TestNumMessages*TestMessageSize), TestBatchSize*TestDepth*TestMessageSize/KERNEL_DATATYPE_SIZE)
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
	afterEachShardCL(t, shard, context)
	fmt.Printf("... done \n")
}

func XTestShardCLReadv3(t *testing.T) {
	fmt.Printf("TestShardCLReadv0: ...\n")
	context, err := NewContextCL("contextcl", KernelCL3)
	if err != nil {
		t.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, TestDepth*TestMessageSize, generateData(TestNumMessages*TestMessageSize), TestNumMessages*TestMessageSize/KERNEL_DATATYPE_SIZE)
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
	afterEachShardCL(t, shard, context)
	fmt.Printf("... done \n")
}

func XBenchmarkShardCLReadv0(b *testing.B) {
	batchSize := 1
	context, err := NewContextCL("contextcl", KernelCL0)
	if err != nil {
		b.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, BenchDepth*BenchMessageSize, generateData(BenchNumMessages*BenchMessageSize), batchSize*context.GetGroupSize())
	if err != nil {
		b.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperBenchmarkShardRead(b, shard, batchSize)
	afterEachShardCL(b, shard, context)
}
