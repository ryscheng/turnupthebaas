package pir

// Shard abstracts out the common interface for ShardCPU, ShardCUDA, and ShardOpenCL.
// A Shard represents an immutable range of data for PIR operations
// Each are backed by a different PIR implementation
// Databases are range partitioned by bucket into Shards.
// Thus, a shard represents an range of `numBuckets` buckets,
// where each bucket is []byte of length `bucketSize`.
// Note: len(data) must equal (numBuckets * bucketSize)
type Shard interface {
	Free() error
	GetName() string
	GetBucketSize() int
	GetNumBuckets() int
	GetData() []byte
	Read(reqs []byte, reqLength int) ([]byte, error)
}
