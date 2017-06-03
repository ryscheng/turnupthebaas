package pircpu

import (
	"fmt"
	"testing"

	"github.com/privacylab/talek/common"
	pt "github.com/privacylab/talek/pir/pirtest"
)

func beforeEach() {
	common.SilenceLoggers()
}

func TestShardCPUCreate(t *testing.T) {
	fmt.Printf("TestShardCPUCreate: ...\n")
	beforeEach()
	// Creating with invalid bucketSize
	shard, err := NewShardCPU("shardcpuv0", 7, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), 0)
	if err == nil {
		t.Fatalf("new ShardCPU should have failed with invalid bucketSize, but didn't return error")
	}
	if shard != nil {
		t.Fatalf("new ShardCPU should have failed with invalid bucketSize, but returned a shard")
	}
	fmt.Printf("... done \n")
}

func TestNewShardInvalidBucketSize(t *testing.T) {
	fmt.Printf("TestShardCPUCreate: ...\n")
	beforeEach()
	// Creating with invalid bucketSize
	shard := NewShard(7, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), "cpu.0")
	if shard != nil {
		t.Fatalf("new ShardCPU should have failed with invalid bucketSize, but returned a shard")
	}
	fmt.Printf("... done \n")
}

func TestNewShardInvalidUserData1(t *testing.T) {
	fmt.Printf("TestShardCPUCreate: ...\n")
	beforeEach()
	// Creating with invalid bucketSize
	shard := NewShard(7, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), "cpu")
	if shard != nil {
		t.Fatalf("new ShardCPU should have failed with invalid user data, but returned a shard")
	}
	fmt.Printf("... done \n")
}

func TestNewShardInvalidUserData2(t *testing.T) {
	fmt.Printf("TestShardCPUCreate: ...\n")
	beforeEach()
	// Creating with invalid bucketSize
	shard := NewShard(7, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), "cpu.cpu")
	if shard != nil {
		t.Fatalf("new ShardCPU should have failed with invalid user data, but returned a shard")
	}
	fmt.Printf("... done \n")
}

func TestShardCPUReadv0(t *testing.T) {
	fmt.Printf("TestShardCPUReadv0: ...\n")
	beforeEach()
	shard := NewShard(pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), "cpu.0")
	if shard == nil {
		t.Fatalf("cannot create new ShardCPU v0\n")
	}
	pt.HelperTestShardRead(t, shard)
	pt.HelperTestClientRead(t, shard)
	pt.AfterEach(t, shard, nil)
	fmt.Printf("... done \n")
}

func TestShardCPUReadv1(t *testing.T) {
	fmt.Printf("TestShardCPUReadv1: ...\n")
	beforeEach()
	shard := NewShard(pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), "cpu.1")
	if shard == nil {
		t.Fatalf("cannot create new ShardCPU v1\n")
	}
	pt.HelperTestShardRead(t, shard)
	pt.HelperTestClientRead(t, shard)
	pt.AfterEach(t, shard, nil)
	fmt.Printf("... done \n")
}

func TestShardCPUReadv2(t *testing.T) {
	fmt.Printf("TestShardCPUReadv2: ...\n")
	beforeEach()
	shard := NewShard(pt.TestDepth*pt.TestMessageSize, pt.GenerateData(pt.TestNumMessages*pt.TestMessageSize), "cpu.2")
	if shard == nil {
		t.Fatalf("cannot create new ShardCPU v2\n")
	}
	pt.HelperTestShardRead(t, shard)
	pt.HelperTestClientRead(t, shard)
	pt.AfterEach(t, shard, nil)
	fmt.Printf("... done \n")
}

func BenchmarkShardCPUReadv0(b *testing.B) {
	//fmt.Printf("BenchmarkShardCPUReadv0 began with N=%d... \n", b.N)
	beforeEach()
	shard := NewShard(pt.BenchDepth*pt.BenchMessageSize, pt.GenerateData(pt.BenchNumMessages*pt.BenchMessageSize), "cpu.0")
	if shard == nil {
		b.Fatalf("cannot create new ShardCPU v0\n")
	}
	pt.HelperBenchmarkShardRead(b, shard, pt.BenchBatchSize)
	pt.AfterEach(b, shard, nil)
}

func BenchmarkShardCPUReadv1(b *testing.B) {
	//fmt.Printf("BenchmarkShardCPUReadv1 began with N=%d... \n", b.N)
	beforeEach()
	shard := NewShard(pt.BenchDepth*pt.BenchMessageSize, pt.GenerateData(pt.BenchNumMessages*pt.BenchMessageSize), "cpu.1")
	if shard == nil {
		b.Fatalf("cannot create new ShardCPU v1\n")
	}
	pt.HelperBenchmarkShardRead(b, shard, pt.BenchBatchSize)
	pt.AfterEach(b, shard, nil)
}

func BenchmarkShardCPUReadv2(b *testing.B) {
	//fmt.Printf("BenchmarkShardCPUReadv2 began with N=%d... \n", b.N)
	beforeEach()
	shard := NewShard(pt.BenchDepth*pt.BenchMessageSize, pt.GenerateData(pt.BenchNumMessages*pt.BenchMessageSize), "cpu.2")
	if shard == nil {
		b.Fatalf("cannot create new ShardCPU v2\n")
	}
	pt.HelperBenchmarkShardRead(b, shard, pt.BenchBatchSize)
	pt.AfterEach(b, shard, nil)
}
