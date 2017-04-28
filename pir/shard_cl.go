package pir

import (
	"fmt"
	"github.com/privacylab/talek/common"
)

// ShardCL represents a read-only shard of the database,
// backed by an OpenCL implementation of PIR
type ShardCL struct {
	// Private State
	log         *common.Logger
	name        string
	bucketSize  int
	numBuckets  int
	data        []byte
	readVersion int
}

// NewShardCL creates a new OpenCL-backed shard
// The data is represented as a flat byte array = append(bucket_1, bucket_2 ... bucket_n)
// Pre-conditions:
// - len(data) must be a multiple of bucketSize
// Returns: the shard, or an error if mismatched size
func NewShardCL(name string, bucketSize int, data []byte, readVersion int) (*ShardCL, error) {
	s := &ShardCL{}
	s.log = common.NewLogger(name)
	s.name = name

	// GetNumBuckets will compute the number of buckets stored in the Shard
	// If len(s.data) is not cleanly divisible by s.bucketSize,
	// returns an error
	if len(data)%bucketSize != 0 {
		return nil, fmt.Errorf("NewShardCL(%v) failed: data(len=%v) not multiple of bucketSize=%v", name, len(data), bucketSize)
	}

	s.bucketSize = bucketSize
	s.numBuckets = (len(data) / bucketSize)
	s.data = data
	return s, nil
}

// Free currently does nothing. ShardCL waits for the go garbage collector
func (s *ShardCL) Free() error {
	return nil
}

// GetName returns the name of the shard
func (s *ShardCL) GetName() string {
	return s.name
}

// GetBucketSize returns the size (in bytes) of a bucket
func (s *ShardCL) GetBucketSize() int {
	return s.bucketSize
}

// GetNumBuckets returns the number of buckets in the shard
func (s *ShardCL) GetNumBuckets() int {
	return s.numBuckets
}

// GetData returns a slice of the data
func (s *ShardCL) GetData() []byte {
	return s.data[:]
}

// Read handles a batch read, where each request is concatentated into `reqs`
//   each request consists of `reqLength` bytes
//	 Note: every request starts on a byte boundary
// Returns: a single byte array where responses are concatenated by the order in `reqs`
//	 each response consists of `s.bucketSize` bytes
func (s *ShardCL) Read(reqs []byte, reqLength int) ([]byte, error) {
	if len(reqs)%reqLength != 0 {
		return nil, fmt.Errorf("ShardCL.Read expects len(reqs)=%d to be a multiple of reqLength=%d", len(reqs), reqLength)
	} else if s.readVersion == 0 {
		return s.read0(reqs, reqLength)
	}

	return nil, fmt.Errorf("ShardCL.Read: invalid readVersion=%d", s.readVersion)
}

func (s *ShardCL) read0(reqs []byte, reqLength int) ([]byte, error) {
	numReqs := len(reqs) / reqLength
	responses := make([]byte, numReqs*s.bucketSize)

	// calculate PIR

	return responses, nil
}
