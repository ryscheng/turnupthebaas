package common

import (
	"time"
)

type GlobalConfig struct {
	NumBuckets    int
	BucketDepth   int
	WindowSize    int
	DataSize      int // Number of bytes
	WriteInterval time.Duration
	ReadInterval  time.Duration
	TrustDomains  []*TrustDomainConfig
}
