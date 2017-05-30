package common

import (
	"testing"
	"time"
)

func testConfig() Config {
	return Config{
		NumBuckets:         8,
		BucketDepth:        2,
		DataSize:           256,
		NumBucketsPerShard: 2,
		NumShardsPerGroup:  2,
		WriteInterval:      time.Minute,
		ReadInterval:       time.Minute,
		MaxLoadFactor:      0.50,
		BloomFalsePositive: 0.01,
	}
}

func TestWindowSize(t *testing.T) {
	conf := testConfig()
	w := conf.WindowSize()
	if w != 8 {
		t.Errorf("TestWindowSize should have been 8")
	}
	conf.MaxLoadFactor = 0.75
	w = conf.WindowSize()
	if w != 12 {
		t.Errorf("TestWindowSize should have been 12")
	}
}

func TestBucketToShardGroup(t *testing.T) {
	conf := testConfig()
	s, g := conf.BucketToShardGroup(0)
	if s != 0 || g != 0 {
		t.Errorf("TestBucketToShardGroup should have returned (0,0)")
	}
	s, g = conf.BucketToShardGroup(7)
	if s != 3 || g != 1 {
		t.Errorf("TestBucketToShardGroup should have returned (0,0)")
	}
	s, g = conf.BucketToShardGroup(8)
	if s != 4 || g != 2 {
		t.Errorf("TestBucketToShardGroup should have returned (0,0)")
	}
}
