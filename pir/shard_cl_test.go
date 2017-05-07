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

func TestShardCLReadv0(t *testing.T) {
	fmt.Printf("TestShardCLReadv0: ...\n")
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
	shard, err := NewShardCL("shardcl", context, TestDepth*TestMessageSize, generateData(TestNumMessages*TestMessageSize), TestBatchSize*TestDepth*TestMessageSize/KernelDataSize)
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
	afterEachShardCL(t, shard, context)
	fmt.Printf("... done \n")
}

func TestShardCLReadv2(t *testing.T) {
	fmt.Printf("TestShardCLReadv2: ...\n")
	context, err := NewContextCL("contextcl", KernelCL2)
	if err != nil {
		t.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, TestDepth*TestMessageSize, generateData(TestNumMessages*TestMessageSize), TestBatchSize*TestDepth*TestMessageSize/KernelDataSize)
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
	afterEachShardCL(t, shard, context)
	fmt.Printf("... done \n")
}

func TestShardCLReadv3(t *testing.T) {
	fmt.Printf("TestShardCLReadv3: ...\n")
	context, err := NewContextCL("contextcl", KernelCL3)
	if err != nil {
		t.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, TestDepth*TestMessageSize, generateData(TestNumMessages*TestMessageSize), TestNumMessages*TestMessageSize/KernelDataSize)
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
	afterEachShardCL(t, shard, context)
	fmt.Printf("... done \n")
}

func BenchmarkShardCLReadv0(b *testing.B) {
	batchSize := 8
	context, err := NewContextCL("contextcl", KernelCL0)
	if err != nil {
		b.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, BenchDepth*BenchMessageSize, generateData(BenchNumMessages*BenchMessageSize), batchSize*context.GetGroupSize())
	if err != nil {
		b.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperBenchmarkShardRead(b, shard, batchSize)
	afterEachShardCL(b, shard, context)
}

func BenchmarkShardCLReadv1(b *testing.B) {
	batchSize := 8
	context, err := NewContextCL("contextcl", KernelCL1)
	if err != nil {
		b.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, BenchDepth*BenchMessageSize, generateData(BenchNumMessages*BenchMessageSize), batchSize*BenchDepth*BenchMessageSize/KernelDataSize)
	if err != nil {
		b.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperBenchmarkShardRead(b, shard, batchSize)
	afterEachShardCL(b, shard, context)
}

func BenchmarkShardCLReadv2(b *testing.B) {
	batchSize := 8
	context, err := NewContextCL("contextcl", KernelCL2)
	if err != nil {
		b.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, BenchDepth*BenchMessageSize, generateData(BenchNumMessages*BenchMessageSize), batchSize*BenchDepth*BenchMessageSize/KernelDataSize)
	if err != nil {
		b.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperBenchmarkShardRead(b, shard, batchSize)
	afterEachShardCL(b, shard, context)
}

func BenchmarkShardCLReadv3(b *testing.B) {
	batchSize := 8
	context, err := NewContextCL("contextcl", KernelCL3)
	if err != nil {
		b.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, BenchDepth*BenchMessageSize, generateData(BenchNumMessages*BenchMessageSize), BenchNumMessages*BenchMessageSize/KernelDataSize)

	if err != nil {
		b.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperBenchmarkShardRead(b, shard, batchSize)
	afterEachShardCL(b, shard, context)
}
