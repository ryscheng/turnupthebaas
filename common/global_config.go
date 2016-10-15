package common

import (
	"time"
)

type GlobalConfig struct {
	NumBuckets    uint32
	BucketDepth   uint32
	WindowSize    uint64
	DataSize      uint32 // Number of bytes
	WriteInterval time.Duration
	ReadInterval  time.Duration
	TrustDomains  []*TrustDomainConfig
}
