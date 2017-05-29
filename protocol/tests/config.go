package tests

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/privacylab/talek/common"
)

func randAddr() string {
	num := rand.Int()
	num %= 100
	num += 9800
	return "localhost:" + strconv.Itoa(num)
}

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
