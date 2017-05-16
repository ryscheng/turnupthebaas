package pirinterface

import "strings"

// Shard abstracts out the common interface for ShardCPU, ShardCUDA, and ShardOpenCL.
// A Shard represents an immutable range of data for PIR operations
// Each are backed by a different PIR implementation
// Databases are range partitioned by bucket into Shards.
// Thus, a shard represents an range of `numBuckets` buckets,
// where each bucket is []byte of length `bucketSize`.
// Note: len(data) must equal (numBuckets * bucketSize)
type Shard interface {
	Free() error
	GetBucketSize() int
	GetNumBuckets() int
	GetData() []byte
	Read(reqs []byte, reqLength int) ([]byte, error)
}

// backings is a static table of the registered / available PIR implementations
var backings map[string]func(int, []byte, string) Shard

// Register allows PIR interfaces to register themselves for later discovery
func Register(prefix string, cons func(int, []byte, string) Shard) {
	if backings == nil {
		backings = make(map[string]func(int, []byte, string) Shard)
	}
	backings[prefix] = cons
}

// GetBacking allows a client to search for available shard constructors
func GetBacking(prefix string) func(int, []byte, string) Shard {
	if backings == nil {
		return nil
	}
	for k, v := range backings {
		if strings.HasPrefix(prefix, k) {
			return v
		}
	}
	return nil
}
