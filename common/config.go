package common

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

// Config is a shared configuration needed by both libtalek and server
// NOTE: all uint64 properties must be a factor of 2
// @todo support other values
type Config struct {
	// How many buckets are in the server?
	NumBuckets uint64
	// How many items are in a bucket?
	BucketDepth uint64
	// How many bytes are in an item?
	DataSize uint64 // Number of bytes
	// Number of buckets in a shard
	NumBucketsPerShard uint64
	// Number of shards in a replica group
	NumShardsPerGroup uint64
	// Minimum period between writes
	WriteInterval time.Duration `json:",string"`
	// Minimum period between reads
	ReadInterval time.Duration `json:",string"`
	// Max fraction of DB capacity that can store messages
	MaxLoadFactor float64
	// False positive rate of interest vectors
	BloomFalsePositive float64

	/** @todo deprecated **/
	// What fraction of items should be removed from the DB when items are removed?
	LoadFactorStep float64
}

// WindowSize is a computed property of Config for how many items are available at a time
func (c *Config) WindowSize() uint64 {
	return uint64(float64(c.NumBuckets*c.BucketDepth) * c.MaxLoadFactor)
}

// BucketToShardGroup outputs the (shard, replica_group) that a bucket is assigned to
func (c *Config) BucketToShardGroup(bucket uint64) (uint64, uint64) {
	shard := bucket / c.NumBucketsPerShard
	group := shard / c.NumShardsPerGroup
	return shard, group
}

// ConfigFromFile restores a JSON file. returns the config on success or nil if
// loading or parsing the file fails.
func ConfigFromFile(file string) *Config {
	configString, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}
	config := new(Config)
	if err := json.Unmarshal(configString, config); err != nil {
		return nil
	}
	return config
}
