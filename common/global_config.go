package common

import (
	"time"
)

type GlobalConfig struct {
	// How many Buckets are in the server?
	NumBuckets         uint64
	// How many items are in a bucket?
	BucketDepth        int
	WindowSize         int
	// How many bytes are in an item?
	DataSize           int // Number of bytes
	// How many read requests should be made of the PIR server at a time?
	ReadBatch          int
	// At what false positive rate should the bloom filter be extended?
	BloomFalsePositive float64
	// At what fraction of DB capacity should items be removed?
	MaxLoadFactor      float32
	// What fraction of items should be removed from the DB when items are removed?
	LoadFactorStep     float32
	WriteInterval      time.Duration
	ReadInterval       time.Duration
	TrustDomains       []*TrustDomainConfig
}
