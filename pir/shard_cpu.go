package pir

import (
	"fmt"
	"github.com/privacylab/talek/common"
)

// ShardCPU represents a read-only shard of the database
// backed by a CPU implementation of PIR
type ShardCPU struct {
	// Private State
	log         *common.Logger
	name        string
	bucketSize  int
	numBuckets  int
	data        []byte
	readVersion int
}

// NewShardCPU creates a new CPU-backed shard
// The data is represented as a flat byte array = append(bucket_1, bucket_2 ... bucket_n)
// Pre-conditions:
// - len(data) must be a multiple of bucketSize
// Returns: the shard, or an error if mismatched size
func NewShardCPU(name string, bucketSize int, data []byte, readVersion int) (*ShardCPU, error) {
	s := &ShardCPU{}
	s.log = common.NewLogger(name)
	s.name = name

	// GetNumBuckets will compute the number of buckets stored in the Shard
	// If len(s.data) is not cleanly divisible by s.bucketSize,
	// returns an error
	if len(data)%bucketSize != 0 {
		return nil, fmt.Errorf("NewShardCPU(%v) failed: data(len=%v) not multiple of bucketSize=%v", name, len(data), bucketSize)
	}

	s.bucketSize = bucketSize
	s.numBuckets = (len(data) / bucketSize)
	s.data = data
	s.readVersion = readVersion
	return s, nil
}

// Free currently does nothing. ShardCPU waits for the go garbage collector
func (s *ShardCPU) Free() error {
	return nil
}

// GetName returns the name of the shard
func (s *ShardCPU) GetName() string {
	return s.name
}

// GetBucketSize returns the size (in bytes) of a bucket
func (s *ShardCPU) GetBucketSize() int {
	return s.bucketSize
}

// GetNumBuckets returns the number of buckets in the shard
func (s *ShardCPU) GetNumBuckets() int {
	return s.numBuckets
}

// GetData returns a slice of the data
func (s *ShardCPU) GetData() []byte {
	return s.data[:]
}

// Insert copies the given byte array into the specified bucket at a given offset
// Returns the number of bytes copied. This will be <len(toCopy) if toCopy is
// Note: This function will only output a warning if it overwrites into the next bucket
/**
func (s *ShardCPU) Insert(bucket int, offset int, toCopy []byte) int {
	index := (bucket * s.bucketSize) + offset
	dst := s.data[index:]
	if len(toCopy) > (s.bucketSize - offset) {
		s.log.Warn.Printf("Shard.Insert overwriting next bucket\n")
	}
	return copy(dst, toCopy)
}
**/

// Read handles a batch read, where each request is concatentated into `reqs`
//   each request consists of `reqLength` bytes
//	 Note: every request starts on a byte boundary
// Returns: a single byte array where responses are concatenated by the order in `reqs`
//	 each response consists of `s.bucketSize` bytes
func (s *ShardCPU) Read(reqs []byte, reqLength int) ([]byte, error) {
	if len(reqs)%reqLength != 0 {
		return nil, fmt.Errorf("ShardCPU.Read expects len(reqs)=%d to be a multiple of reqLength=%d", len(reqs), reqLength)
	} else if s.readVersion == 0 {
		return s.read0(reqs, reqLength)
	} else if s.readVersion == 1 {
		return s.read1(reqs, reqLength)
	}

	return nil, fmt.Errorf("ShardCPU.Read: invalid readVersion=%d", s.readVersion)
}

func (s *ShardCPU) read0(reqs []byte, reqLength int) ([]byte, error) {
	numReqs := len(reqs) / reqLength
	responses := make([]byte, numReqs*s.bucketSize)

	// calculate PIR
	// Note: Much better for outer loop to be reqIndex, not bucketIndex
	for reqIndex := 0; reqIndex < numReqs; reqIndex++ {
		for bucketIndex := 0; bucketIndex < s.numBuckets; bucketIndex++ {
			reqByte := reqs[reqIndex*reqLength+(bucketIndex/8)]
			if reqByte&(byte(1)<<uint(bucketIndex%8)) != 0 {
				bucket := s.data[(bucketIndex * s.bucketSize):]
				response := responses[(reqIndex * s.bucketSize):]
				for offset := 0; offset < s.bucketSize; offset++ {
					response[offset] ^= bucket[offset]
				}
			}
		}
	}

	return responses, nil
}

func (s *ShardCPU) read1(reqs []byte, reqLength int) ([]byte, error) {
	numReqs := len(reqs) / reqLength
	responses := make([]byte, numReqs*s.bucketSize)

	// calculate PIR
	for reqIndex := 0; reqIndex < numReqs; reqIndex++ {
		reqOffset := reqIndex * reqLength
		respOffset := reqIndex * s.bucketSize
		for bucketIndex := 0; bucketIndex < s.numBuckets; bucketIndex++ {
			reqByte := reqs[reqOffset+(bucketIndex/8)]
			if reqByte&(byte(1)<<uint(bucketIndex%8)) != 0 {
				bucketOffset := bucketIndex * s.bucketSize
				bucket := s.data[bucketOffset:(bucketOffset + s.bucketSize)]
				response := responses[respOffset:(respOffset + s.bucketSize)]
				xorWords(response, response, bucket)
			}
		}
	}

	return responses, nil
}
