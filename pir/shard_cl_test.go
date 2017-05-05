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
	fmt.Printf("TestShardCLReadv0: ...\n")
	context, err := NewContextCL("contextcl", KernelSource0)
	if err != nil {
		t.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, TestDepth*TestMessageSize, generateData(TestNumMessages*TestMessageSize), 0)
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
	afterEachShardCL(t, shard, context)
	fmt.Printf("... done \n")
}

func XTestShardCLReadv1(t *testing.T) {
	fmt.Printf("TestShardCLReadv1: ...\n")
	context, err := NewContextCL("contextcl", KernelSource1)
	if err != nil {
		t.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, TestDepth*TestMessageSize, generateData(TestNumMessages*TestMessageSize), 1)
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
	afterEachShardCL(t, shard, context)
	fmt.Printf("... done \n")
}

func XTestShardCLReadv2(t *testing.T) {
	fmt.Printf("TestShardCLReadv1: ...\n")
	context, err := NewContextCL("contextcl", KernelSourceX)
	if err != nil {
		t.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, TestDepth*TestMessageSize, generateData(TestNumMessages*TestMessageSize), 2)
	if err != nil {
		t.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperTestShardRead(t, shard)
	afterEachShardCL(t, shard, context)
	fmt.Printf("... done \n")
}

func BenchmarkShardCLReadv2(b *testing.B) {
	context, err := NewContextCL("contextcl", KernelSourceX)
	if err != nil {
		b.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	shard, err := NewShardCL("shardcl", context, BenchDepth*BenchMessageSize, generateData(BenchNumMessages*BenchMessageSize), 2)
	if err != nil {
		b.Fatalf("cannot create new ShardCL: error=%v\n", err)
	}
	HelperBenchmarkShardRead(b, shard, 32)
	afterEachShardCL(b, shard, context)
}
