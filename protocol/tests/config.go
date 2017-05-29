package tests

import (
	"time"

	"github.com/privacylab/talek/common"
)

const testAddr = "localhost:9876"

func testConfig() common.Config {
	return common.Config{
		NumBuckets:         8,
		BucketDepth:        2,
		DataSize:           256,
		BloomFalsePositive: 0.01,
		WriteInterval:      time.Minute,
		ReadInterval:       time.Minute,
		MaxLoadFactor:      0.50,
	}
}
