package common

import (
	"time"
)

type GlobalConfig struct {
	NumBuckets    uint32
	DataSize      uint32
	WriteInterval time.Duration
	ReadInterval  time.Duration
	TrustDomains  []*TrustDomainConfig
}
