package replica

import (
	"time"

	"github.com/privacylab/talek/common"
)

/********************************
 *** HELPER FUNCTIONS
 ********************************/

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
