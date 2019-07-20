package common

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

// Config is a shared configuration needed by both libtalek and server
type Config struct {
	// How many buckets are in the server?
	NumBuckets uint64
	// How many items are in a bucket?
	BucketDepth uint64
	// How many bytes are in an item?
	DataSize uint64 // Number of bytes
	// False positive rate of interest vectors
	BloomFalsePositive float64
	// Minimum period between writes
	WriteInterval time.Duration `json:",string"`
	// Minimum period between reads
	ReadInterval time.Duration `json:",string"`
	// On what multiple of write intervals will new interest vectors be computed
	InterestMultiple uint64
	// Base seed for hashing interest vectors
	InterestSeed int64
	// Max fraction of DB capacity that can store messages
	MaxLoadFactor float64

	/** @todo remove below **/
	// What fraction of items should be removed from the DB when items are removed?
	LoadFactorStep float64
}

// WindowSize is a computed property of Config for how many items are available at a time
func (cc *Config) WindowSize() uint64 {
	return uint64(float64(cc.NumBuckets*cc.BucketDepth) * cc.MaxLoadFactor)
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
