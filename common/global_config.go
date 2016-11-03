package common

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

type GlobalConfig struct {
	// How many Buckets are in the server?
	NumBuckets uint64
	// How many items are in a bucket?
	BucketDepth int
	// How many bytes are in an item?
	DataSize int // Number of bytes
	// How many read requests should be made of the PIR server at a time?
	ReadBatch int
	// At what false positive rate should the bloom filter be extended?
	BloomFalsePositive float64
	// At what fraction of DB capacity should items be removed?
	MaxLoadFactor float32
	// What fraction of items should be removed from the DB when items are removed?
	LoadFactorStep float32
	WriteInterval  time.Duration
	ReadInterval   time.Duration
	TrustDomains   []*TrustDomainConfig `json:"-"`
}

func (g *GlobalConfig) WindowSize() int {
	return int(float32(int(g.NumBuckets) * g.BucketDepth) * g.MaxLoadFactor)
}

// Load configuration from a JSON file. returns the config on success or nil if
// loading or parsing the file fails.
func GlobalConfigFromFile(file string) *GlobalConfig {
	configString, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}
	config := new(GlobalConfig)
	if err := json.Unmarshal(configString, config); err != nil {
		return nil
	}
	return config
}
