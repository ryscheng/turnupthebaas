package common

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

type CommonConfig struct {
	// How many Buckets are in the server?
	NumBuckets uint64
	// How many items are in a bucket?
	BucketDepth int
	// How many bytes are in an item?
	DataSize int // Number of bytes
	// At what false positive rate should the bloom filter be extended?
	BloomFalsePositive float64
	// At what fraction of DB capacity should items be removed?
	MaxLoadFactor float32
	// What fraction of items should be removed from the DB when items are removed?
	LoadFactorStep float32

	// How often should pending writes be applied to the database.
	WriteInterval  time.Duration

	// What is the minimum interval with which reads should occur.
	ReadInterval   time.Duration

	// Where are the different servers?
	TrustDomains   []*TrustDomainConfig `json:"-"`
}

func (cc *CommonConfig) WindowSize() int {
	return int(float32(int(cc.NumBuckets) * cc.BucketDepth) * cc.MaxLoadFactor)
}

// Load configuration from a JSON file. returns the config on success or nil if
// loading or parsing the file fails.
func CommonConfigFromFile(file string) *CommonConfig {
	configString, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}
	config := new(CommonConfig)
	if err := json.Unmarshal(configString, config); err != nil {
		return nil
	}
	return config
}
